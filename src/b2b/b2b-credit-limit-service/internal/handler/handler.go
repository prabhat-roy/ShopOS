package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopos/b2b-credit-limit-service/internal/domain"
	"github.com/shopos/b2b-credit-limit-service/internal/service"
)

// Handler bundles the HTTP mux and service layer.
type Handler struct {
	mux *http.ServeMux
	svc service.Servicer
}

// New wires up all routes and returns a ready Handler.
func New(svc service.Servicer) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: svc}
	h.routes()
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) routes() {
	h.mux.HandleFunc("GET /healthz", h.healthz)
	h.mux.HandleFunc("POST /credit-limits", h.setLimit)
	h.mux.HandleFunc("GET /credit-limits/{id}", h.getLimit)
	h.mux.HandleFunc("GET /credit-limits/org/{orgId}", h.getByOrg)
	h.mux.HandleFunc("POST /credit-limits/org/{orgId}/utilize", h.utilize)
	h.mux.HandleFunc("POST /credit-limits/org/{orgId}/payment", h.payment)
	h.mux.HandleFunc("PATCH /credit-limits/org/{orgId}/limit", h.adjustLimit)
	h.mux.HandleFunc("POST /credit-limits/org/{orgId}/suspend", h.suspend)
	h.mux.HandleFunc("POST /credit-limits/org/{orgId}/review", h.review)
	h.mux.HandleFunc("GET /credit-limits/org/{orgId}/check", h.checkAvailability)
	h.mux.HandleFunc("GET /credit-limits/org/{orgId}/history", h.history)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// setLimit handles POST /credit-limits — creates or updates the credit limit for an org.
func (h *Handler) setLimit(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrgID    string  `json:"org_id"`
		Limit    float64 `json:"credit_limit"`
		Currency string  `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	orgID, err := uuid.Parse(body.OrgID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid org_id")
		return
	}
	cl, err := h.svc.SetCreditLimit(orgID, body.Limit, body.Currency)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cl)
}

// getLimit handles GET /credit-limits/{id}.
func (h *Handler) getLimit(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	cl, err := h.svc.GetCreditLimit(id)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// getByOrg handles GET /credit-limits/org/{orgId}.
func (h *Handler) getByOrg(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	cl, err := h.svc.GetByOrg(orgID)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// utilize handles POST /credit-limits/org/{orgId}/utilize.
func (h *Handler) utilize(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	var body struct {
		Amount    float64 `json:"amount"`
		Reference string  `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	cl, err := h.svc.UtilizeCredit(orgID, body.Amount, body.Reference)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// payment handles POST /credit-limits/org/{orgId}/payment.
func (h *Handler) payment(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	var body struct {
		Amount    float64 `json:"amount"`
		Reference string  `json:"reference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	cl, err := h.svc.MakePayment(orgID, body.Amount, body.Reference)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// adjustLimit handles PATCH /credit-limits/org/{orgId}/limit.
func (h *Handler) adjustLimit(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	var body struct {
		NewLimit float64 `json:"new_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	cl, err := h.svc.AdjustLimit(orgID, body.NewLimit)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// suspend handles POST /credit-limits/org/{orgId}/suspend.
func (h *Handler) suspend(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	if err := h.svc.SuspendOrg(orgID); err != nil {
		writeCreditError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// review handles POST /credit-limits/org/{orgId}/review.
func (h *Handler) review(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	var body struct {
		RiskScore int `json:"risk_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %s", err))
		return
	}
	cl, err := h.svc.ReviewCredit(orgID, body.RiskScore)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// checkAvailability handles GET /credit-limits/org/{orgId}/check?amount=.
func (h *Handler) checkAvailability(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	rawAmount := r.URL.Query().Get("amount")
	if rawAmount == "" {
		writeError(w, http.StatusBadRequest, "amount query parameter is required")
		return
	}
	amount, err := strconv.ParseFloat(rawAmount, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "amount must be a valid number")
		return
	}
	check, err := h.svc.CheckAvailability(orgID, amount)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, check)
}

// history handles GET /credit-limits/org/{orgId}/history.
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	orgID, ok := parseUUID(w, r, "orgId")
	if !ok {
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	txs, err := h.svc.GetCreditHistory(orgID, limit)
	if err != nil {
		writeCreditError(w, err)
		return
	}
	if txs == nil {
		txs = []*domain.CreditTransaction{}
	}
	writeJSON(w, http.StatusOK, txs)
}

// --- helpers ---

func parseUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	id, err := uuid.Parse(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid %s: %s", key, raw))
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}

func writeCreditError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInsufficientCredit):
		writeError(w, http.StatusPaymentRequired, err.Error())
	case errors.Is(err, domain.ErrAccountSuspended):
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}
