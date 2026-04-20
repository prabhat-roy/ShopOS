package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/webhook-service/internal/domain"
)

// Servicer is the subset of service.WebhookService used by HTTP handlers.
type Servicer interface {
	Create(ctx context.Context, req *domain.CreateWebhookRequest) (*domain.Webhook, error)
	Get(ctx context.Context, id string) (*domain.Webhook, error)
	List(ctx context.Context, ownerID string) ([]*domain.Webhook, error)
	Update(ctx context.Context, id string, req *domain.UpdateWebhookRequest) (*domain.Webhook, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	svc Servicer
}

func New(svc Servicer) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /webhooks", h.create)
	mux.HandleFunc("GET /webhooks", h.list)
	mux.HandleFunc("GET /webhooks/{id}", h.get)
	mux.HandleFunc("PATCH /webhooks/{id}", h.update)
	mux.HandleFunc("DELETE /webhooks/{id}", h.delete)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	hook, err := h.svc.Create(r.Context(), &req)
	if errors.Is(err, domain.ErrInvalidURL) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, hook)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner_id")
	hooks, err := h.svc.List(r.Context(), ownerID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if hooks == nil {
		hooks = []*domain.Webhook{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": hooks, "count": len(hooks)})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	hook, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, hook)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req domain.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	hook, err := h.svc.Update(r.Context(), id, &req)
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, domain.ErrInvalidURL) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, hook)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.svc.Delete(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
