package handler

import (
	"net/http"

	"github.com/shopos/partner-bff/internal/service"
)

type CatalogHandler struct{ svc service.CatalogService }

func NewCatalogHandler(svc service.CatalogService) *CatalogHandler { return &CatalogHandler{svc: svc} }

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.svc.ListProducts(r.Context(),
		queryInt(q.Get("page"), 1),
		queryInt(q.Get("page_size"), 20),
		q.Get("category_id"),
	)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, list)
}

func (h *CatalogHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "product id is required"})
		return
	}
	p, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, p)
}

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.svc.ListCategories(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": cats})
}
