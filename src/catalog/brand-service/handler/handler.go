package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/shopos/brand-service/domain"
	"github.com/shopos/brand-service/service"
)

// Servicer is the interface the handler depends on (facilitates mocking in tests).
type Servicer interface {
	Create(req service.CreateRequest) (*domain.Brand, error)
	GetByID(id string) (*domain.Brand, error)
	GetBySlug(slug string) (*domain.Brand, error)
	List(activeOnly bool) ([]*domain.Brand, error)
	Update(id string, req service.UpdateRequest) (*domain.Brand, error)
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
	h.mux.HandleFunc("/brands", h.brands)
	h.mux.HandleFunc("/brands/", h.brandByPath)

	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---- route handlers ---------------------------------------------------------

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// brands handles POST /brands and GET /brands.
func (h *Handler) brands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createBrand(w, r)
	case http.MethodGet:
		h.listBrands(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// brandByPath dispatches /brands/{id}, /brands/slug/{slug}, PATCH and DELETE.
func (h *Handler) brandByPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/brands/")
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// /brands/slug/{slug}
	if strings.HasPrefix(path, "slug/") {
		slug := strings.TrimPrefix(path, "slug/")
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.getBrandBySlug(w, r, slug)
		return
	}

	// /brands/{id}
	id := path
	switch r.Method {
	case http.MethodGet:
		h.getBrandByID(w, r, id)
	case http.MethodPatch:
		h.updateBrand(w, r, id)
	case http.MethodDelete:
		h.deleteBrand(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) createBrand(w http.ResponseWriter, r *http.Request) {
	var req service.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	b, err := h.svc.Create(req)
	if err != nil {
		if errors.Is(err, domain.ErrSlugTaken) {
			writeError(w, http.StatusConflict, "slug already taken")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) listBrands(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	brands, err := h.svc.List(activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if brands == nil {
		brands = []*domain.Brand{}
	}
	writeJSON(w, http.StatusOK, brands)
}

func (h *Handler) getBrandByID(w http.ResponseWriter, r *http.Request, id string) {
	b, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) getBrandBySlug(w http.ResponseWriter, r *http.Request, slug string) {
	b, err := h.svc.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) updateBrand(w http.ResponseWriter, r *http.Request, id string) {
	var req service.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	b, err := h.svc.Update(id, req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		if errors.Is(err, domain.ErrSlugTaken) {
			writeError(w, http.StatusConflict, "slug already taken")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) deleteBrand(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
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
