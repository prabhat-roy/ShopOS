package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/gift-card-service/domain"
	"github.com/shopos/gift-card-service/service"
)

// Handler holds the HTTP mux and the service dependency.
type Handler struct {
	mux *http.ServeMux
	svc *service.Service
}

// New wires up all routes and returns the Handler.
func New(svc *service.Service) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.mux.HandleFunc("GET /healthz", h.healthz)
	h.mux.HandleFunc("POST /gift-cards", h.issue)
	h.mux.HandleFunc("GET /gift-cards/{code}", h.getCard)
	h.mux.HandleFunc("GET /gift-cards/{code}/balance", h.checkBalance)
	h.mux.HandleFunc("POST /gift-cards/{code}/redeem", h.redeem)
	h.mux.HandleFunc("DELETE /gift-cards/{code}", h.deactivate)
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// healthz returns a simple liveness response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// issue creates and returns a new gift card.
// Body: {"initial_balance": <float>, "currency": "<string>", "issued_to": "<string>", "expires_at": "<RFC3339|omit>"}
func (h *Handler) issue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InitialBalance float64  `json:"initial_balance"`
		Currency       string   `json:"currency"`
		IssuedTo       string   `json:"issued_to"`
		ExpiresAt      *RFCTime `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}
	if req.InitialBalance <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("initial_balance must be positive"))
		return
	}

	var expiresAt *RFCTime
	expiresAt = req.ExpiresAt

	card, err := h.svc.IssueCard(req.InitialBalance, req.Currency, req.IssuedTo, timePtr(expiresAt))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, card)
}

// getCard returns a gift card by code.
func (h *Handler) getCard(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	card, err := h.svc.GetCard(code)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, card)
}

// checkBalance returns the current balance metadata for a card.
func (h *Handler) checkBalance(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	card, err := h.svc.CheckBalance(code)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"balance":    card.CurrentBalance,
		"currency":   card.Currency,
		"expires_at": card.ExpiresAt,
	})
}

// redeem applies a redemption to the gift card.
// Body: {"order_id": "<string>", "amount": <float>}
func (h *Handler) redeem(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	var req struct {
		OrderID string  `json:"order_id"`
		Amount  float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}
	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("amount must be positive"))
		return
	}

	rec, err := h.svc.Redeem(code, req.OrderID, req.Amount)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

// deactivate deactivates the gift card identified by code.
func (h *Handler) deactivate(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if err := h.svc.Deactivate(code); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// writeError maps domain errors to HTTP status codes.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorBody(err.Error()))
	case errors.Is(err, domain.ErrInsufficientBalance):
		writeJSON(w, http.StatusPaymentRequired, errorBody(err.Error()))
	case errors.Is(err, domain.ErrCardExpired):
		writeJSON(w, http.StatusUnprocessableEntity, errorBody(err.Error()))
	case errors.Is(err, domain.ErrCardInactive):
		writeJSON(w, http.StatusUnprocessableEntity, errorBody(err.Error()))
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
