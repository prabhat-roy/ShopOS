package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/shopos/worker-job-queue/internal/domain"
)

// Servicer is the business-logic interface the Handler depends on.
// Any type that satisfies this interface (real service or mock) can be
// injected, making the handler straightforward to test.
type Servicer interface {
	Enqueue(ctx context.Context, req *domain.EnqueueRequest) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListDead(ctx context.Context, queue string) ([]*domain.Job, error)
	Retry(ctx context.Context, queue, id string) (*domain.Job, error)
}

// Handler holds the HTTP multiplexer and a reference to the service layer.
type Handler struct {
	mux     *http.ServeMux
	service Servicer
}

// New wires up all routes and returns the Handler.
//
// Routes registered:
//
//	POST   /queues/{queue}/jobs           → enqueue (201 Created)
//	GET    /queues/{queue}/jobs/{id}      → get job
//	GET    /queues/{queue}/dead           → list dead jobs
//	POST   /queues/{queue}/dead/{id}/retry → retry dead job (201 Created)
//	GET    /healthz                        → {"status":"ok"}
func New(svc Servicer) *Handler {
	h := &Handler{
		mux:     http.NewServeMux(),
		service: svc,
	}
	h.registerRoutes()
	return h
}

// ServeHTTP implements http.Handler so Handler can be passed directly to
// http.ListenAndServe or httptest.NewServer.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// registerRoutes attaches all route patterns to the internal ServeMux.
// Go 1.22+ path-value syntax ({param}) is used; values are retrieved with
// r.PathValue("param").
func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("GET /healthz", h.healthz)
	h.mux.HandleFunc("POST /queues/{queue}/jobs", h.enqueue)
	h.mux.HandleFunc("GET /queues/{queue}/jobs/{id}", h.getJob)
	h.mux.HandleFunc("GET /queues/{queue}/dead", h.listDead)
	h.mux.HandleFunc("POST /queues/{queue}/dead/{id}/retry", h.retryDead)
}

// ---------------------------------------------------------------------------
// Route handlers
// ---------------------------------------------------------------------------

// healthz responds with {"status":"ok"} and HTTP 200.
func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// enqueue decodes an EnqueueRequest from the request body, calls
// Servicer.Enqueue, and responds 201 with the created Job.
func (h *Handler) enqueue(w http.ResponseWriter, r *http.Request) {
	queue := r.PathValue("queue")

	var req domain.EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	// Override queue from URL path so callers cannot spoof the queue name via
	// the JSON body.
	req.Queue = queue

	job, err := h.service.Enqueue(r.Context(), &req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

// getJob fetches a single job by ID.
func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	job, err := h.service.GetJob(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// listDead returns dead-letter queue contents for the named queue.
func (h *Handler) listDead(w http.ResponseWriter, r *http.Request) {
	queue := r.PathValue("queue")

	jobs, err := h.service.ListDead(r.Context(), queue)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// Always return a JSON array, never null.
	if jobs == nil {
		jobs = []*domain.Job{}
	}

	writeJSON(w, http.StatusOK, jobs)
}

// retryDead re-enqueues a dead job and responds 201 with the new Job.
func (h *Handler) retryDead(w http.ResponseWriter, r *http.Request) {
	queue := r.PathValue("queue")
	id := r.PathValue("id")

	job, err := h.service.Retry(r.Context(), queue, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// errorResponse is the JSON envelope for all error responses.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// writeError writes a JSON error body with the given status and message.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// handleServiceError maps well-known domain errors to HTTP status codes.
func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrQueueEmpty):
		writeError(w, http.StatusNoContent, err.Error())
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
