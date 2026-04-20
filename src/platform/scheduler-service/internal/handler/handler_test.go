package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/scheduler-service/internal/domain"
	"github.com/shopos/scheduler-service/internal/handler"
)

type mockSvc struct {
	job  *domain.Job
	jobs []*domain.Job
	runs []*domain.JobRun
	err  error
}

func (m *mockSvc) CreateJob(_ context.Context, _ *domain.CreateJobRequest) (*domain.Job, error) {
	return m.job, m.err
}
func (m *mockSvc) GetJob(_ context.Context, _ string) (*domain.Job, error) {
	return m.job, m.err
}
func (m *mockSvc) ListJobs(_ context.Context) ([]*domain.Job, error) {
	return m.jobs, m.err
}
func (m *mockSvc) UpdateJob(_ context.Context, _ string, _ *domain.UpdateJobRequest) (*domain.Job, error) {
	return m.job, m.err
}
func (m *mockSvc) DeleteJob(_ context.Context, _ string) error {
	return m.err
}
func (m *mockSvc) ListRuns(_ context.Context, _ string, _ int) ([]*domain.JobRun, error) {
	return m.runs, m.err
}

var _ handler.Servicer = (*mockSvc)(nil)

func build(svc handler.Servicer) http.Handler {
	mux := http.NewServeMux()
	handler.New(svc).Register(mux)
	return mux
}

func newJob() *domain.Job {
	now := time.Now()
	return &domain.Job{
		ID:         "job-1",
		Name:       "nightly-report",
		CronExpr:   "0 2 * * *",
		HTTPMethod: "POST",
		HTTPURL:    "https://internal/run",
		Status:     domain.StatusEnabled,
		NextRunAt:  now.Add(time.Hour),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func TestHealth(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateJob(t *testing.T) {
	svc := &mockSvc{job: newJob()}
	h := build(svc)

	body, _ := json.Marshal(domain.CreateJobRequest{
		Name:       "nightly-report",
		CronExpr:   "0 2 * * *",
		HTTPMethod: "POST",
		HTTPURL:    "https://internal/run",
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp domain.Job
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ID != "job-1" {
		t.Errorf("expected job-1, got %s", resp.ID)
	}
}

func TestCreateJobBadBody(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateJobInvalidCron(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrInvalidCron})
	body, _ := json.Marshal(domain.CreateJobRequest{CronExpr: "bad-cron"})
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListJobs(t *testing.T) {
	svc := &mockSvc{jobs: []*domain.Job{newJob(), newJob()}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["count"].(float64) != 2 {
		t.Errorf("expected 2, got %v", resp["count"])
	}
}

func TestGetJob(t *testing.T) {
	h := build(&mockSvc{job: newJob()})
	req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
	req.SetPathValue("id", "job-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetJobNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/jobs/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateJob(t *testing.T) {
	svc := &mockSvc{job: newJob()}
	h := build(svc)

	newExpr := "0 3 * * *"
	body, _ := json.Marshal(domain.UpdateJobRequest{CronExpr: &newExpr})
	req := httptest.NewRequest(http.MethodPatch, "/jobs/job-1", bytes.NewReader(body))
	req.SetPathValue("id", "job-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteJob(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/jobs/job-1", nil)
	req.SetPathValue("id", "job-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteJobNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodDelete, "/jobs/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestListRuns(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{runs: []*domain.JobRun{
		{ID: "run-1", JobID: "job-1", Status: domain.RunSuccess, StartedAt: now, FinishedAt: now},
	}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/jobs/job-1/runs?limit=10", nil)
	req.SetPathValue("id", "job-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("expected 1 run, got %v", resp["count"])
	}
}
