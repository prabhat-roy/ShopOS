package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/shopos/tenant-service/internal/domain"
)

// Servicer is the business-logic interface required by Handler.
type Servicer interface {
	Create(ctx context.Context, req domain.CreateTenantRequest) (*domain.Tenant, error)
	Get(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	List(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error)
	Update(ctx context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error)
	Delete(ctx context.Context, id string) error
	GetSettings(ctx context.Context, id string) (map[string]string, error)
	UpdateSettings(ctx context.Context, id string, settings map[string]string) (*domain.Tenant, error)
}

// Handler wires all HTTP routes for the tenant service.
type Handler struct {
	svc Servicer
	mux *http.ServeMux
}

// New creates a Handler and registers all routes on the provided ServeMux.
func New(svc Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.registerRoutes()
	return h
}

// ServeHTTP satisfies http.Handler so Handler can be used directly with http.Server.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("GET /healthz", h.handleHealthz)
	h.mux.HandleFunc("POST /tenants", h.handleCreateTenant)
	h.mux.HandleFunc("GET /tenants", h.handleListTenants)
	// More-specific patterns must come before the general {id} pattern.
	h.mux.HandleFunc("GET /tenants/slug/{slug}", h.handleGetTenantBySlug)
	h.mux.HandleFunc("GET /tenants/{id}", h.handleGetTenant)
	h.mux.HandleFunc("PATCH /tenants/{id}", h.handleUpdateTenant)
	h.mux.HandleFunc("DELETE /tenants/{id}", h.handleDeleteTenant)
	h.mux.HandleFunc("GET /tenants/{id}/settings", h.handleGetSettings)
	h.mux.HandleFunc("PUT /tenants/{id}/settings", h.handleUpdateSettings)
}

// — route handlers —

func (h *Handler) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.svc.Create(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tenant)
}

func (h *Handler) handleListTenants(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	limit := parseIntParam(q.Get("limit"), 20)
	offset := parseIntParam(q.Get("offset"), 0)

	tenants, err := h.svc.List(r.Context(), status, limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if tenants == nil {
		tenants = []*domain.Tenant{}
	}
	writeJSON(w, http.StatusOK, tenants)
}

func (h *Handler) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tenant, err := h.svc.Get(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (h *Handler) handleGetTenantBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tenant, err := h.svc.GetBySlug(r.Context(), slug)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (h *Handler) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req domain.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (h *Handler) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	settings, err := h.svc.GetSettings(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var settings map[string]string
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenant, err := h.svc.UpdateSettings(r.Context(), id, settings)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

// — helpers —

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrSlugTaken):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func parseIntParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
