package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/rate-limiter-service/internal/domain"
)

// Servicer is the interface the HTTP handlers depend on.
type Servicer interface {
	CreatePolicy(ctx context.Context, req *domain.CreatePolicyRequest) (*domain.Policy, error)
	GetPolicy(ctx context.Context, id string) (*domain.Policy, error)
	ListPolicies(ctx context.Context) ([]*domain.Policy, error)
	UpdatePolicy(ctx context.Context, id string, req *domain.UpdatePolicyRequest) (*domain.Policy, error)
	DeletePolicy(ctx context.Context, id string) error
	Check(ctx context.Context, req domain.CheckRequest) (*domain.CheckResponse, error)
}

// Handler holds all HTTP handlers.
type Handler struct{ svc Servicer }

func New(svc Servicer) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /policies", h.listPolicies)
	mux.HandleFunc("POST /policies", h.createPolicy)
	mux.HandleFunc("GET /policies/{id}", h.getPolicy)
	mux.HandleFunc("PATCH /policies/{id}", h.updatePolicy)
	mux.HandleFunc("DELETE /policies/{id}", h.deletePolicy)
	mux.HandleFunc("POST /check", h.check)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := h.svc.ListPolicies(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": policies, "total": len(policies)})
}

func (h *Handler) getPolicy(w http.ResponseWriter, r *http.Request) {
	p, err := h.svc.GetPolicy(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) createPolicy(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	p, err := h.svc.CreatePolicy(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) updatePolicy(w http.ResponseWriter, r *http.Request) {
	var req domain.UpdatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	p, err := h.svc.UpdatePolicy(r.Context(), r.PathValue("id"), &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *Handler) deletePolicy(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeletePolicy(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) check(w http.ResponseWriter, r *http.Request) {
	var req domain.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	resp, err := h.svc.Check(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	code := http.StatusOK
	if !resp.Allowed {
		code = http.StatusTooManyRequests
	}
	writeJSON(w, code, resp)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrRateLimited):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
