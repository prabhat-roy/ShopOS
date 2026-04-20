package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shopos/partner-bff/internal/service"
)

type InventoryHandler struct{ svc service.InventoryService }

func NewInventoryHandler(svc service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc: svc}
}

func (h *InventoryHandler) GetStock(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "product id is required"})
		return
	}
	stock, err := h.svc.GetStock(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, stock)
}

func (h *InventoryHandler) GetBulkStock(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductIDs []string `json:"product_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.ProductIDs) == 0 {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "product_ids array is required"})
		return
	}
	if len(body.ProductIDs) > 100 {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "maximum 100 product_ids per request"})
		return
	}
	levels, err := h.svc.GetBulkStock(r.Context(), body.ProductIDs)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": levels})
}
