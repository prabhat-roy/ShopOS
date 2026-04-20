package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/feature-flag-service/internal/domain"
)

// Servicer is the interface the handlers depend on.
type Servicer interface {
	GetFlag(ctx context.Context, key string) (*domain.Flag, error)
	ListFlags(ctx context.Context) ([]*domain.Flag, error)
	CreateFlag(ctx context.Context, req *domain.CreateFlagRequest) (*domain.Flag, error)
	UpdateFlag(ctx context.Context, id string, req *domain.UpdateFlagRequest) (*domain.Flag, error)
	DeleteFlag(ctx context.Context, id string) error
	Evaluate(ctx context.Context, req domain.EvalRequest) (bool, error)
}

// Handler holds all HTTP handlers for the feature-flag service.
type Handler struct{ svc Servicer }

func New(svc Servicer) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /flags", h.listFlags)
	mux.HandleFunc("POST /flags", h.createFlag)
	mux.HandleFunc("GET /flags/{key}", h.getFlag)
	mux.HandleFunc("PATCH /flags/{id}", h.updateFlag)
	mux.HandleFunc("DELETE /flags/{id}", h.deleteFlag)
	mux.HandleFunc("POST /flags/evaluate", h.evaluate)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.svc.ListFlags(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": flags, "total": len(flags)})
}

func (h *Handler) getFlag(w http.ResponseWriter, r *http.Request) {
	flag, err := h.svc.GetFlag(r.Context(), r.PathValue("key"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *Handler) createFlag(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	flag, err := h.svc.CreateFlag(r.Context(), &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, flag)
}

func (h *Handler) updateFlag(w http.ResponseWriter, r *http.Request) {
	var req domain.UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	flag, err := h.svc.UpdateFlag(r.Context(), r.PathValue("id"), &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, flag)
}

func (h *Handler) deleteFlag(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteFlag(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) evaluate(w http.ResponseWriter, r *http.Request) {
	var req domain.EvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "key is required"})
		return
	}
	enabled, err := h.svc.Evaluate(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"key": req.Key, "enabled": enabled})
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
	case errors.Is(err, domain.ErrAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	case errors.Is(err, domain.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

// GRPCEvaluate is the gRPC entry point stub (wired after proto compilation).
func (h *Handler) GRPCEvaluate(ctx context.Context, key, userID string, ctxMap map[string]string) (bool, error) {
	return h.svc.Evaluate(ctx, domain.EvalRequest{Key: key, UserID: userID, Context: ctxMap})
}
