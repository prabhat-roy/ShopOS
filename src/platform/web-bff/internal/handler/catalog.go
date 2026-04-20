package handler

import (
	"net/http"
	"strconv"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/service"
)

type CatalogHandler struct {
	svc service.CatalogService
}

func NewCatalogHandler(svc service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := &domain.ListProductsRequest{
		Page:       queryInt(q.Get("page"), 1),
		PageSize:   queryInt(q.Get("page_size"), 20),
		CategoryID: q.Get("category_id"),
		Sort:       q.Get("sort"),
	}
	if v := q.Get("min_price"); v != "" {
		req.MinPrice, _ = strconv.ParseFloat(v, 64)
	}
	if v := q.Get("max_price"); v != "" {
		req.MaxPrice, _ = strconv.ParseFloat(v, 64)
	}
	list, err := h.svc.ListProducts(r.Context(), req)
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
	product, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, product)
}

func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.ListCategories(r.Context())
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, map[string]any{"items": categories})
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		JSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter q is required"})
		return
	}
	page := queryInt(r.URL.Query().Get("page"), 1)
	pageSize := queryInt(r.URL.Query().Get("page_size"), 20)

	results, err := h.svc.Search(r.Context(), q, page, pageSize)
	if err != nil {
		Error(w, err)
		return
	}
	JSON(w, http.StatusOK, results)
}

func queryInt(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return fallback
}
