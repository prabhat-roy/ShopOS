package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/shopos/bundle-service/domain"
)

// Servicer is the application logic contract consumed by the HTTP layer.
type Servicer interface {
	Create(ctx context.Context, b *domain.Bundle) (*domain.Bundle, error)
	GetByID(ctx context.Context, id string) (*domain.Bundle, error)
	List(ctx context.Context, activeOnly bool) ([]*domain.Bundle, error)
	Update(ctx context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error)
	Delete(ctx context.Context, id string) error
}

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	svc Servicer
}

// New returns a Handler wired to the given Servicer.
func New(svc Servicer) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes attaches all routes to the provided ServeMux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", h.healthz)
	mux.HandleFunc("/bundles", h.bundlesCollection)
	mux.HandleFunc("/bundles/", h.bundleItem)
}

// -----------------------------------------------------------------------
// Health
// -----------------------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// -----------------------------------------------------------------------
// Collection routes  POST /bundles  GET /bundles
// -----------------------------------------------------------------------

func (h *Handler) bundlesCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createBundle(w, r)
	case http.MethodGet:
		h.listBundles(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /bundles
func (h *Handler) createBundle(w http.ResponseWriter, r *http.Request) {
	var b domain.Bundle
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if b.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if b.Currency == "" {
		b.Currency = "USD"
	}
	if b.Items == nil {
		b.Items = []domain.BundleItem{}
	}
	b.Active = true

	created, err := h.svc.Create(r.Context(), &b)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// GET /bundles?active=true
func (h *Handler) listBundles(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	bundles, err := h.svc.List(r.Context(), activeOnly)
	if err != nil {
		handleError(w, err)
		return
	}
	if bundles == nil {
		bundles = []*domain.Bundle{}
	}
	writeJSON(w, http.StatusOK, bundles)
}

// -----------------------------------------------------------------------
// Item routes  GET /bundles/{id}  PATCH /bundles/{id}  DELETE /bundles/{id}
// -----------------------------------------------------------------------

func (h *Handler) bundleItem(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/bundles/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getBundle(w, r, id)
	case http.MethodPatch:
		h.updateBundle(w, r, id)
	case http.MethodDelete:
		h.deleteBundle(w, r, id)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /bundles/{id}
func (h *Handler) getBundle(w http.ResponseWriter, r *http.Request, id string) {
	b, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// PATCH /bundles/{id}
func (h *Handler) updateBundle(w http.ResponseWriter, r *http.Request, id string) {
	var patch domain.Bundle
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	updated, err := h.svc.Update(r.Context(), id, &patch)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DELETE /bundles/{id}
func (h *Handler) deleteBundle(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		log.Printf("internal error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
