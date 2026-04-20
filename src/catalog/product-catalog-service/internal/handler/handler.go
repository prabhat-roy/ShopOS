package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/shopos/product-catalog-service/internal/domain"
	"github.com/shopos/product-catalog-service/internal/service"
)

// Servicer defines the methods the handler needs from the service layer.
// Defined as an interface to allow mocking in handler tests.
type Servicer interface {
	CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	ListProducts(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error)
	UpdateProduct(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id string) error
}

// Handler wires HTTP routes to the service layer.
type Handler struct {
	svc    Servicer
	logger *slog.Logger
}

// New creates a Handler and registers all routes on the provided mux.
func New(svc Servicer, logger *slog.Logger) http.Handler {
	h := &Handler{svc: svc, logger: logger}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("POST /products", h.createProduct)
	mux.HandleFunc("GET /products", h.listProducts)
	mux.HandleFunc("GET /products/sku/{sku}", h.getProductBySKU)
	mux.HandleFunc("GET /products/{id}", h.getProduct)
	mux.HandleFunc("PATCH /products/{id}", h.updateProduct)
	mux.HandleFunc("DELETE /products/{id}", h.deleteProduct)

	return mux
}

// ---------------------------------------------------------------------------
// Route handlers
// ---------------------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.CreateProduct(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, product)
}

func (h *Handler) listProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := &domain.ListProductsRequest{
		CategoryID: q.Get("category_id"),
		BrandID:    q.Get("brand_id"),
		Status:     domain.ProductStatus(q.Get("status")),
		Limit:      parseIntQuery(q.Get("limit"), 20),
		Offset:     parseIntQuery(q.Get("offset"), 0),
	}

	if minPrice := q.Get("min_price"); minPrice != "" {
		if v, err := strconv.ParseFloat(minPrice, 64); err == nil {
			req.MinPrice = v
		}
	}
	if maxPrice := q.Get("max_price"); maxPrice != "" {
		if v, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			req.MaxPrice = v
		}
	}

	list, err := h.svc.ListProducts(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	product, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) getProductBySKU(w http.ResponseWriter, r *http.Request) {
	sku := r.PathValue("sku")
	product, err := h.svc.GetBySKU(r.Context(), sku)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req domain.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.svc.UpdateProduct(r.Context(), id, &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *Handler) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.DeleteProduct(r.Context(), id); err != nil {
		h.handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Error mapping
// ---------------------------------------------------------------------------

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "product not found")
	case errors.Is(err, domain.ErrDuplicateSKU):
		writeError(w, http.StatusConflict, "SKU already exists")
	default:
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Error())
			return
		}
		h.logger.Error("internal error", "err", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Nothing we can do here — headers already sent.
		return
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func parseIntQuery(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
