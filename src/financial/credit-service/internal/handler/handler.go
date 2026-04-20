package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/shopos/credit-service/internal/domain"
	"github.com/shopos/credit-service/internal/service"
)

// Handler holds dependencies for HTTP route handlers.
type Handler struct {
	svc service.Servicer
}

// New creates a Handler and registers all routes on mux.
func New(svc service.Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes registers all HTTP endpoints on the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/credit-accounts", h.creditAccounts)
	mux.HandleFunc("/credit-accounts/", h.creditAccountsRouter)
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("handler: writeJSON encode error: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// ---- /healthz ---------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- /credit-accounts -------------------------------------------------------

// creditAccounts handles POST /credit-accounts.
func (h *Handler) creditAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	h.createAccount(w, r)
}

type createAccountRequest struct {
	CustomerID  string  `json:"customer_id"`
	CreditLimit float64 `json:"credit_limit"`
	Currency    string  `json:"currency"`
}

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	customerID, err := parseUUID(req.CustomerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer_id UUID")
		return
	}
	if req.CreditLimit < 0 {
		writeError(w, http.StatusBadRequest, "credit_limit must be non-negative")
		return
	}
	acc, err := h.svc.CreateCreditAccount(customerID, req.CreditLimit, req.Currency)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

// ---- /credit-accounts/{id}[/...] router ------------------------------------

// creditAccountsRouter dispatches sub-paths under /credit-accounts/.
func (h *Handler) creditAccountsRouter(w http.ResponseWriter, r *http.Request) {
	// Strip leading slash and split: ["credit-accounts", segment, ...]
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// parts[0] == "credit-accounts"

	if len(parts) < 2 || parts[1] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// /credit-accounts/customer/{customerId}
	if parts[1] == "customer" {
		if len(parts) < 3 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		h.getByCustomerID(w, r, parts[2])
		return
	}

	accountID, err := parseUUID(parts[1])
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	// No sub-resource — GET /credit-accounts/{id}
	if len(parts) == 2 {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getAccount(w, r, accountID)
		return
	}

	// Sub-resource
	switch parts[2] {
	case "charge":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.chargeCredit(w, r, accountID)
	case "payment":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.makePayment(w, r, accountID)
	case "limit":
		if r.Method != http.MethodPatch {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.adjustLimit(w, r, accountID)
	case "suspend":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.suspendAccount(w, r, accountID)
	case "close":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.closeAccount(w, r, accountID)
	case "transactions":
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.listTransactions(w, r, accountID)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

// GET /credit-accounts/{id}
func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	acc, err := h.svc.GetCreditAccount(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

// GET /credit-accounts/customer/{customerId}
func (h *Handler) getByCustomerID(w http.ResponseWriter, r *http.Request, rawID string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	customerID, err := parseUUID(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid customer_id UUID")
		return
	}
	acc, err := h.svc.GetByCustomerID(customerID)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

// POST /credit-accounts/{id}/charge
type chargeRequest struct {
	Amount      float64 `json:"amount"`
	Reference   string  `json:"reference"`
	Description string  `json:"description"`
}

func (h *Handler) chargeCredit(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req chargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	tx, err := h.svc.ChargeCredit(id, req.Amount, req.Reference, req.Description)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if errors.Is(err, domain.ErrAccountInactive) {
		writeError(w, http.StatusUnprocessableEntity, "account is not active")
		return
	}
	if errors.Is(err, domain.ErrInsufficientCredit) {
		writeError(w, http.StatusUnprocessableEntity, "insufficient available credit")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

// POST /credit-accounts/{id}/payment
type paymentRequest struct {
	Amount    float64 `json:"amount"`
	Reference string  `json:"reference"`
}

func (h *Handler) makePayment(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req paymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	tx, err := h.svc.MakePayment(id, req.Amount, req.Reference)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if errors.Is(err, domain.ErrAccountInactive) {
		writeError(w, http.StatusUnprocessableEntity, "account is closed")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

// PATCH /credit-accounts/{id}/limit
type limitRequest struct {
	NewLimit float64 `json:"new_limit"`
}

func (h *Handler) adjustLimit(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req limitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.NewLimit < 0 {
		writeError(w, http.StatusBadRequest, "new_limit must be non-negative")
		return
	}
	acc, err := h.svc.AdjustCreditLimit(id, req.NewLimit)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if errors.Is(err, domain.ErrAccountInactive) {
		writeError(w, http.StatusUnprocessableEntity, "account is closed")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

// POST /credit-accounts/{id}/suspend
func (h *Handler) suspendAccount(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := h.svc.SuspendAccount(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /credit-accounts/{id}/close
func (h *Handler) closeAccount(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := h.svc.CloseAccount(id)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /credit-accounts/{id}/transactions
func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	txs, err := h.svc.GetTransactionHistory(id, limit)
	if errors.Is(err, domain.ErrNotFound) {
		writeError(w, http.StatusNotFound, "account not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if txs == nil {
		txs = []domain.CreditTransaction{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": txs})
}
