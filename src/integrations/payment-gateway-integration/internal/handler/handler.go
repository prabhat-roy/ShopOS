package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shopos/payment-gateway-integration/internal/domain"
	"github.com/shopos/payment-gateway-integration/internal/service"
)

// Handler holds all HTTP handler dependencies.
type Handler struct {
	svc *service.Servicer
}

// New returns an initialised Handler.
func New(svc *service.Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches all routes to mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/payments/charge", h.createCharge)
	mux.HandleFunc("/payments/refund", h.createRefund)
	mux.HandleFunc("/payments/refunds/", h.getRefund)
	mux.HandleFunc("/payments/", h.getPayment)
	mux.HandleFunc("/payments", h.listPayments)
	mux.HandleFunc("/gateways", h.listGateways)
}

// healthz returns service health.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// createCharge handles POST /payments/charge
func (h *Handler) createCharge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req domain.ChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.CreateCharge(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// getPayment handles GET /payments/{id}
func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/payments/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "payment id required")
		return
	}

	pi, err := h.svc.GetPayment(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pi)
}

// listPayments handles GET /payments?orderId=&customerId=
func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	orderID := q.Get("orderId")
	customerID := q.Get("customerId")

	if orderID == "" && customerID == "" {
		writeError(w, http.StatusBadRequest, "orderId or customerId query parameter required")
		return
	}

	var intents interface{}
	if orderID != "" {
		intents = h.svc.ListOrderPayments(orderID)
	} else {
		intents = h.svc.ListCustomerPayments(customerID)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"payments": intents})
}

// createRefund handles POST /payments/refund
func (h *Handler) createRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req domain.RefundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.CreateRefund(req)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

// getRefund handles GET /payments/refunds/{refundId}
func (h *Handler) getRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	refundID := strings.TrimPrefix(r.URL.Path, "/payments/refunds/")
	if refundID == "" {
		writeError(w, http.StatusBadRequest, "refund id required")
		return
	}

	ref, err := h.svc.GetRefund(refundID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ref)
}

// listGateways handles GET /gateways
func (h *Handler) listGateways(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"gateways": h.svc.GetSupportedGateways(),
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
