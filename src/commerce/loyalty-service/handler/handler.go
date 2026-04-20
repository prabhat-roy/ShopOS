package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/shopos/loyalty-service/domain"
	"github.com/shopos/loyalty-service/service"
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
	h.mux.HandleFunc("GET /loyalty/{customerID}", h.getAccount)
	h.mux.HandleFunc("POST /loyalty/{customerID}/earn", h.earn)
	h.mux.HandleFunc("POST /loyalty/{customerID}/redeem", h.redeem)
	h.mux.HandleFunc("GET /loyalty/{customerID}/transactions", h.listTransactions)
	return h
}

// ServeHTTP implements http.Handler so Handler can be used directly as the server handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// healthz returns a simple liveness response.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// getAccount returns the loyalty account and current balance for a customer.
func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")
	acc, err := h.svc.GetAccount(customerID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"customer_id": acc.CustomerID,
		"points":      acc.Points,
		"tier":        acc.TierName,
		"created_at":  acc.CreatedAt,
		"updated_at":  acc.UpdatedAt,
	})
}

// earn processes a point-earning event.
// Body: {"points": <int>, "order_id": "<string>", "description": "<string>"}
func (h *Handler) earn(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")

	var req struct {
		Points      int64  `json:"points"`
		OrderID     string `json:"order_id"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}
	if req.Points <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("points must be positive"))
		return
	}

	txn, err := h.svc.EarnPoints(customerID, req.Points, req.OrderID, req.Description)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, txn)
}

// redeem processes a points-redemption request.
// Body: {"points": <int>, "order_id": "<string>"}
func (h *Handler) redeem(w http.ResponseWriter, r *http.Request) {
	customerID := r.PathValue("customerID")

	var req struct {
		Points  int64  `json:"points"`
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid request body"))
		return
	}
	if req.Points <= 0 {
		writeJSON(w, http.StatusBadRequest, errorBody("points must be positive"))
		return
	}

	txn, dollarValue, err := h.svc.RedeemPoints(customerID, req.Points, req.OrderID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"transaction": txn,
		"dollar_value": dollarValue,
	})
}

// listTransactions returns recent transactions for a customer.
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
		txns = []domain.PointTransaction{}
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
	case errors.Is(err, domain.ErrInsufficientPoints):
		writeJSON(w, http.StatusPaymentRequired, errorBody(err.Error()))
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}

func errorBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
