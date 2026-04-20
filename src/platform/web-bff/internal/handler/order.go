package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/service"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r.URL.Query().Get("page"), 1)
	pageSize := queryInt(r.URL.Query().Get("page_size"), 10)

	orders, err := h.svc.ListOrders(r.Context(), userID(r), page, pageSize)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": orders})
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("id")
	order, err := h.svc.GetOrder(r.Context(), orderID, userID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, order)
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req domain.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.AddressID == "" || req.PaymentID == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "address_id and payment_id are required"})
		return
	}
	order, err := h.svc.PlaceOrder(r.Context(), userID(r), &req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, order)
}
