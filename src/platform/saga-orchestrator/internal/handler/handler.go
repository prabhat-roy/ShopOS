package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/saga-orchestrator/internal/domain"
)

// Servicer is the interface the HTTP handlers depend on.
type Servicer interface {
	Start(ctx context.Context, req domain.StartSagaRequest) (*domain.Saga, error)
	GetSaga(ctx context.Context, id string) (*domain.Saga, error)
	GetSagaByOrder(ctx context.Context, orderID string) (*domain.Saga, error)
}

// Handler holds all HTTP handlers.
type Handler struct{ svc Servicer }

func New(svc Servicer) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /sagas", h.startSaga)
	mux.HandleFunc("GET /sagas/{id}", h.getSaga)
	mux.HandleFunc("GET /sagas/order/{orderID}", h.getSagaByOrder)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) startSaga(w http.ResponseWriter, r *http.Request) {
	var req domain.StartSagaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	saga, err := h.svc.Start(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, saga)
}

func (h *Handler) getSaga(w http.ResponseWriter, r *http.Request) {
	saga, err := h.svc.GetSaga(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saga)
}

func (h *Handler) getSagaByOrder(w http.ResponseWriter, r *http.Request) {
	saga, err := h.svc.GetSagaByOrder(r.Context(), r.PathValue("orderID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, saga)
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
	case errors.Is(err, domain.ErrInvalidState):
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}
