package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/shopos/wallet-service/domain"
	"github.com/shopos/wallet-service/service"
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
	h.mux.HandleFunc("GET /wallets/{customerID}", h.getWallet)
	h.mux.HandleFunc("POST /wallets/{customerID}/credit", h.credit)
	h.mux.HandleFunc("POST /wallets/{customerID}/debit", h.debit)
	h.mux.HandleFunc("GET /wallets/{customerID}/transactions", h.listTransactions)
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

// getWallet returns the wallet and current balance for a customer.
func (h *Handler) getWallet(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	wallet, err := h.svc.GetOrCreateWallet(customerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":          wallet.ID,
		"customer_id": wallet.CustomerID,
		"balance":     wallet.Balance,
		"currency":    wallet.Currency,
		"created_at":  wallet.CreatedAt,
		"updated_at":  wallet.UpdatedAt,
	})
}

// credit adds funds to the customer's wallet.
// Body: {"amount": <float>, "reference": "<string>", "description": "<string>"}
func (h *Handler) credit(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")

	var req struct {
		Amount      float64 `json:"amount"`
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}
	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("amount must be positive"))
		return
	}

	txn, err := h.svc.Credit(customerID, req.Amount, req.Reference, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

// debit removes funds from the customer's wallet.
// Body: {"amount": <float>, "reference": "<string>", "description": "<string>"}
// Returns HTTP 402 Payment Required when balance is insufficient.
func (h *Handler) debit(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")

	var req struct {
		Amount      float64 `json:"amount"`
		Reference   string  `json:"reference"`
		Description string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}
	if req.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("amount must be positive"))
		return
	}

	txn, err := h.svc.Debit(customerID, req.Amount, req.Reference, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, txn)
}

// listTransactions returns recent transactions for a customer's wallet.
// Optional query param: limit (default 50, max 100).
func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	txns, err := h.svc.GetTransactions(customerID, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	if txns == nil {
		txns = []domain.WalletTransaction{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": txns})
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
	case errors.Is(err, domain.ErrInsufficientFunds):
		writeJSON(w, http.StatusPaymentRequired, errorBody(err.Error()))
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
