package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/service"
)

type OrderHandler struct{ svc service.OrderService }

func NewOrderHandler(svc service.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	orders, err := h.svc.ListOrders(r.Context(), partnerID(r),
		queryInt(q.Get("page"), 1),
		queryInt(q.Get("page_size"), 20),
	)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": orders})
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	order, err := h.svc.GetOrder(r.Context(), r.PathValue("id"), partnerID(r))
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
	if len(req.Items) == 0 || req.AddressID == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "items and address_id are required"})
		return
	}
	order, err := h.svc.PlaceOrder(r.Context(), partnerID(r), &req)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusCreated, order)
}
