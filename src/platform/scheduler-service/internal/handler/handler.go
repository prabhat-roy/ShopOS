package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/shopos/scheduler-service/internal/domain"
)

// Servicer is the subset of SchedulerService used by HTTP handlers.
type Servicer interface {
	CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.Job, error)
	GetJob(ctx context.Context, id string) (*domain.Job, error)
	ListJobs(ctx context.Context) ([]*domain.Job, error)
	UpdateJob(ctx context.Context, id string, req *domain.UpdateJobRequest) (*domain.Job, error)
	DeleteJob(ctx context.Context, id string) error
	ListRuns(ctx context.Context, jobID string, limit int) ([]*domain.JobRun, error)
}

type Handler struct {
	svc Servicer
}

func New(svc Servicer) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /jobs", h.createJob)
	mux.HandleFunc("GET /jobs", h.listJobs)
	mux.HandleFunc("GET /jobs/{id}", h.getJob)
	mux.HandleFunc("PATCH /jobs/{id}", h.updateJob)
	mux.HandleFunc("DELETE /jobs/{id}", h.deleteJob)
	mux.HandleFunc("GET /jobs/{id}/runs", h.listRuns)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createJob(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	job, err := h.svc.CreateJob(r.Context(), &req)
	if errors.Is(err, domain.ErrInvalidCron) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.svc.ListJobs(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if jobs == nil {
		jobs = []*domain.Job{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": jobs, "count": len(jobs)})
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := h.svc.GetJob(r.Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) updateJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req domain.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	job, err := h.svc.UpdateJob(r.Context(), id, &req)
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, domain.ErrInvalidCron) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (h *Handler) deleteJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.svc.DeleteJob(r.Context(), id)
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

func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	runs, err := h.svc.ListRuns(r.Context(), id, limit)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if runs == nil {
		runs = []*domain.JobRun{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": runs, "count": len(runs)})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
