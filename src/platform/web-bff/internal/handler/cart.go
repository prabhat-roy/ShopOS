package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/service"
)

type CartHandler struct {
	svc service.CartService
}

func NewCartHandler(svc service.CartService) *CartHandler {
	return &CartHandler{svc: svc}
}

func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	cart, err := h.svc.GetCart(r.Context(), userID(r))
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req domain.AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.ProductID == "" || req.Quantity <= 0 {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "product_id and quantity > 0 are required"})
		return
	}
	cart, err := h.svc.AddItem(r.Context(), userID(r), req.ProductID, req.Quantity)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemID := r.PathValue("itemId")
	var req domain.UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Quantity <= 0 {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "quantity must be > 0"})
		return
	}
	cart, err := h.svc.UpdateItem(r.Context(), userID(r), itemID, req.Quantity)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID := r.PathValue("itemId")
	cart, err := h.svc.RemoveItem(r.Context(), userID(r), itemID)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, cart)
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.ClearCart(r.Context(), userID(r)); err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "cart cleared"})
}
