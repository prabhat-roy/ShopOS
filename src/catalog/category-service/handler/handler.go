package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/category-service/domain"
	"github.com/shopos/category-service/service"
)

// Servicer is the interface the handler depends on (facilitates mocking in tests).
type Servicer interface {
	Create(req service.CreateRequest) (*domain.Category, error)
	GetByID(id string) (*domain.Category, error)
	GetBySlug(slug string) (*domain.Category, error)
	List(parentID *string, activeOnly bool) ([]*domain.Category, error)
	Update(id string, req service.UpdateRequest) (*domain.Category, error)
	Delete(id string) error
}

// Handler holds the HTTP mux and the service dependency.
type Handler struct {
	mux *http.ServeMux
	svc Servicer
}

// New wires up all routes and returns a ready Handler.
func New(svc Servicer) *Handler {
	h := &Handler{mux: http.NewServeMux(), svc: svc}

	h.mux.HandleFunc("/healthz", h.healthz)
	h.mux.HandleFunc("/categories", h.categories)
	// Routes with path parameters are dispatched manually below.
	h.mux.HandleFunc("/categories/", h.categoryByPath)

	return h
}

// ServeHTTP implements http.Handler so the Handler can be passed directly to http.ListenAndServe.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---- route handlers ---------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// categories handles POST /categories and GET /categories.
func (h *Handler) categories(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createCategory(w, r)
	case http.MethodGet:
		h.listCategories(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// categoryByPath dispatches /categories/{id}, /categories/slug/{slug}, PATCH and DELETE.
func (h *Handler) categoryByPath(w http.ResponseWriter, r *http.Request) {
	// Strip leading "/categories/"
	path := strings.TrimPrefix(r.URL.Path, "/categories/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// /categories/slug/{slug}
	if strings.HasPrefix(path, "slug/") {
		slug := strings.TrimPrefix(path, "slug/")
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getCategoryBySlug(w, r, slug)
		return
	}

	// /categories/{id}
	id := path
	switch r.Method {
	case http.MethodGet:
		h.getCategoryByID(w, r, id)
	case http.MethodPatch:
		h.updateCategory(w, r, id)
	case http.MethodDelete:
		h.deleteCategory(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var req service.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	cat, err := h.svc.Create(req)
	if err != nil {
		if errors.Is(err, domain.ErrSlugTaken) {
			writeError(w, http.StatusConflict, "slug already taken")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var parentID *string
	if p := q.Get("parent_id"); p != "" {
		parentID = &p
	}

	activeOnly := false
	if q.Get("active") == "true" {
		activeOnly = true
	}

	cats, err := h.svc.List(parentID, activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if cats == nil {
		cats = []*domain.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *Handler) getCategoryByID(w http.ResponseWriter, r *http.Request, id string) {
	cat, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *Handler) getCategoryBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	cat, err := h.svc.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request, id string) {
	var req service.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	cat, err := h.svc.Update(id, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		if errors.Is(err, domain.ErrSlugTaken) {
			writeError(w, http.StatusConflict, "slug already taken")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cat)
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "category not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
