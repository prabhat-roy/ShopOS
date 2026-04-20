package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shopos/event-replay-service/internal/domain"
)

// Servicer is the interface the handler layer depends on.
type Servicer interface {
	CreateJob(ctx context.Context, req *domain.CreateReplayRequest) (*domain.ReplayJob, error)
	GetJob(ctx context.Context, id string) (*domain.ReplayJob, error)
	ListJobs(ctx context.Context) ([]*domain.ReplayJob, error)
	CancelJob(ctx context.Context, id string) error
	RunJob(ctx context.Context, id string) error
}

// Handler wires HTTP routes to service calls.
type Handler struct {
	svc Servicer
	mux *http.ServeMux
}

// New builds a Handler and registers all routes on a fresh ServeMux.
func New(svc Servicer) *Handler {
	h := &Handler{svc: svc, mux: http.NewServeMux()}
	h.registerRoutes()
	return h
}

// ServeHTTP satisfies http.Handler so the Handler can be used directly with
// http.ListenAndServe.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("GET /healthz", h.healthz)
	h.mux.HandleFunc("POST /replays", h.createReplay)
	h.mux.HandleFunc("GET /replays", h.listReplays)
	h.mux.HandleFunc("GET /replays/{id}", h.getReplay)
	h.mux.HandleFunc("POST /replays/{id}/run", h.runReplay)
	h.mux.HandleFunc("POST /replays/{id}/cancel", h.cancelReplay)
}

// GET /healthz
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /replays → 202 Accepted
func (h *Handler) createReplay(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid request body"))
		return
	}

	job, err := h.svc.CreateJob(r.Context(), &req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

// GET /replays
func (h *Handler) listReplays(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.svc.ListJobs(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}

	// Return an empty array rather than null when there are no jobs.
	if jobs == nil {
		jobs = []*domain.ReplayJob{}
	}
	writeJSON(w, http.StatusOK, jobs)
}

// GET /replays/{id}
func (h *Handler) getReplay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	job, err := h.svc.GetJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errBody("replay job not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// POST /replays/{id}/run → 202 Accepted
func (h *Handler) runReplay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Verify the job exists before starting the run.
	if _, err := h.svc.GetJob(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errBody("replay job not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}

	// Run asynchronously so we can return 202 immediately.
	go func() {
		_ = h.svc.RunJob(context.Background(), id)
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted", "id": id})
}

// POST /replays/{id}/cancel → 204 No Content
func (h *Handler) cancelReplay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.svc.CancelJob(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errBody("replay job not found"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----------------------------------------------------------------

type errorResponse struct {
	Error string `json:"error"`
}

func errBody(msg string) errorResponse {
	return errorResponse{Error: msg}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
