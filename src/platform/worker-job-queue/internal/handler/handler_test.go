package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/worker-job-queue/internal/domain"
	"github.com/shopos/worker-job-queue/internal/handler"
)

// ---------------------------------------------------------------------------
// mockServicer — in-memory implementation of handler.Servicer
// ---------------------------------------------------------------------------

type mockServicer struct {
	jobs map[string]*domain.Job // keyed by job ID

	// Optional per-call overrides for controlled error injection.
	enqueueErr error
	getJobErr  error
	listDeadFn func(ctx context.Context, queue string) ([]*domain.Job, error)
	retryErr   error
}

func newMockServicer() *mockServicer {
	return &mockServicer{jobs: make(map[string]*domain.Job)}
}

func (m *mockServicer) Enqueue(ctx context.Context, req *domain.EnqueueRequest) (*domain.Job, error) {
	if m.enqueueErr != nil {
		return nil, m.enqueueErr
	}
	job := &domain.Job{
		ID:          "job-001",
		Queue:       req.Queue,
		Priority:    req.Priority,
		Payload:     req.Payload,
		CallbackURL: req.CallbackURL,
		MaxRetries:  req.MaxRetries,
		Status:      domain.StatusPending,
	}
	if job.Priority == "" {
		job.Priority = domain.PriorityNormal
	}
	m.jobs[job.ID] = job
	return job, nil
}

func (m *mockServicer) GetJob(ctx context.Context, id string) (*domain.Job, error) {
	if m.getJobErr != nil {
		return nil, m.getJobErr
	}
	job, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return job, nil
}

func (m *mockServicer) ListDead(ctx context.Context, queue string) ([]*domain.Job, error) {
	if m.listDeadFn != nil {
		return m.listDeadFn(ctx, queue)
	}
	var out []*domain.Job
	for _, j := range m.jobs {
		if j.Queue == queue && j.Status == domain.StatusDead {
			out = append(out, j)
		}
	}
	return out, nil
}

func (m *mockServicer) Retry(ctx context.Context, queue, id string) (*domain.Job, error) {
	if m.retryErr != nil {
		return nil, m.retryErr
	}
	original, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	newJob := &domain.Job{
		ID:          "job-retry-001",
		Queue:       original.Queue,
		Priority:    original.Priority,
		Payload:     original.Payload,
		CallbackURL: original.CallbackURL,
		MaxRetries:  original.MaxRetries,
		Status:      domain.StatusPending,
	}
	m.jobs[newJob.ID] = newJob
	return newJob, nil
}

// compile-time assertion: mockServicer must satisfy handler.Servicer
var _ handler.Servicer = (*mockServicer)(nil)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newRequest builds an HTTP request, optionally setting Go 1.22 path values
// via req.SetPathValue so that r.PathValue() works inside the handler without
// needing a live mux routing cycle.
func newRequest(method, target string, body []byte, pathValues map[string]string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	for k, v := range pathValues {
		req.SetPathValue(k, v)
	}
	return req
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(dst); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests — GET /healthz
// ---------------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	svc := newMockServicer()
	h := handler.New(svc)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	decodeBody(t, w, &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

// ---------------------------------------------------------------------------
// Tests — POST /queues/{queue}/jobs  (enqueue)
// ---------------------------------------------------------------------------

func TestEnqueue_Success(t *testing.T) {
	svc := newMockServicer()
	h := handler.New(svc)

	body := domain.EnqueueRequest{
		Priority:    domain.PriorityHigh,
		Payload:     []byte(`{"key":"value"}`),
		CallbackURL: "http://localhost/cb",
		MaxRetries:  5,
	}
	bodyBytes, _ := json.Marshal(body)

	req := newRequest(http.MethodPost, "/queues/emails/jobs", bodyBytes, map[string]string{"queue": "emails"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var job domain.Job
	decodeBody(t, w, &job)
	if job.ID == "" {
		t.Error("expected non-empty job ID in response")
	}
	if job.Queue != "emails" {
		t.Errorf("expected queue=emails, got %q", job.Queue)
	}
}

func TestEnqueue_InvalidBody(t *testing.T) {
	svc := newMockServicer()
	h := handler.New(svc)

	req := newRequest(http.MethodPost, "/queues/emails/jobs", []byte("not-json"), map[string]string{"queue": "emails"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestEnqueue_ServiceError(t *testing.T) {
	svc := newMockServicer()
	svc.enqueueErr = errors.New("callback_url must not be empty")
	h := handler.New(svc)

	body := domain.EnqueueRequest{Priority: domain.PriorityNormal}
	bodyBytes, _ := json.Marshal(body)

	req := newRequest(http.MethodPost, "/queues/tasks/jobs", bodyBytes, map[string]string{"queue": "tasks"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests — GET /queues/{queue}/jobs/{id}  (getJob)
// ---------------------------------------------------------------------------

func TestGetJob_Found(t *testing.T) {
	svc := newMockServicer()
	// Pre-populate a job directly in the mock store
	svc.jobs["abc-123"] = &domain.Job{
		ID:     "abc-123",
		Queue:  "orders",
		Status: domain.StatusCompleted,
	}
	h := handler.New(svc)

	req := newRequest(http.MethodGet, "/queues/orders/jobs/abc-123", nil, map[string]string{
		"queue": "orders",
		"id":    "abc-123",
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var job domain.Job
	decodeBody(t, w, &job)
	if job.ID != "abc-123" {
		t.Errorf("expected job ID abc-123, got %q", job.ID)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	svc := newMockServicer()
	h := handler.New(svc)

	req := newRequest(http.MethodGet, "/queues/orders/jobs/missing", nil, map[string]string{
		"queue": "orders",
		"id":    "missing",
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetJob_ServiceError(t *testing.T) {
	svc := newMockServicer()
	svc.getJobErr = errors.New("redis unavailable")
	h := handler.New(svc)

	req := newRequest(http.MethodGet, "/queues/orders/jobs/x", nil, map[string]string{
		"queue": "orders",
		"id":    "x",
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for generic service error, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests — GET /queues/{queue}/dead  (listDead)
// ---------------------------------------------------------------------------

func TestListDead_ReturnsArray(t *testing.T) {
	svc := newMockServicer()
	svc.jobs["d1"] = &domain.Job{ID: "d1", Queue: "alerts", Status: domain.StatusDead}
	svc.jobs["d2"] = &domain.Job{ID: "d2", Queue: "alerts", Status: domain.StatusDead}
	h := handler.New(svc)

	req := newRequest(http.MethodGet, "/queues/alerts/dead", nil, map[string]string{"queue": "alerts"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var jobs []*domain.Job
	decodeBody(t, w, &jobs)
	if len(jobs) != 2 {
		t.Errorf("expected 2 dead jobs, got %d", len(jobs))
	}
}

func TestListDead_EmptyQueueReturnsEmptyArray(t *testing.T) {
	svc := newMockServicer()
	h := handler.New(svc)

	req := newRequest(http.MethodGet, "/queues/empty/dead", nil, map[string]string{"queue": "empty"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var jobs []*domain.Job
	decodeBody(t, w, &jobs)
	if len(jobs) != 0 {
		t.Errorf("expected empty array, got %d jobs", len(jobs))
	}
}

func TestListDead_ServiceError(t *testing.T) {
	svc := newMockServicer()
	svc.listDeadFn = func(_ context.Context, _ string) ([]*domain.Job, error) {
		return nil, errors.New("storage error")
	}
	h := handler.New(svc)

	req := newRequest(http.MethodGet, "/queues/broken/dead", nil, map[string]string{"queue": "broken"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests — POST /queues/{queue}/dead/{id}/retry  (retryDead)
// ---------------------------------------------------------------------------

func TestRetryDead_Success(t *testing.T) {
	svc := newMockServicer()
	svc.jobs["dead-job-1"] = &domain.Job{
		ID:          "dead-job-1",
		Queue:       "tasks",
		Priority:    domain.PriorityNormal,
		Status:      domain.StatusDead,
		CallbackURL: "http://localhost/cb",
		MaxRetries:  3,
	}
	h := handler.New(svc)

	req := newRequest(http.MethodPost, "/queues/tasks/dead/dead-job-1/retry", nil, map[string]string{
		"queue": "tasks",
		"id":    "dead-job-1",
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var job domain.Job
	decodeBody(t, w, &job)
	if job.Status != domain.StatusPending {
		t.Errorf("expected new job status=pending, got %q", job.Status)
	}
}

func TestRetryDead_NotFound(t *testing.T) {
	svc := newMockServicer()
	svc.retryErr = domain.ErrNotFound
	h := handler.New(svc)

	req := newRequest(http.MethodPost, "/queues/tasks/dead/ghost/retry", nil, map[string]string{
		"queue": "tasks",
		"id":    "ghost",
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRetryDead_ServiceError(t *testing.T) {
	svc := newMockServicer()
	svc.retryErr = errors.New("cannot retry job with status \"pending\"")
	h := handler.New(svc)

	req := newRequest(http.MethodPost, "/queues/tasks/dead/live-job/retry", nil, map[string]string{
		"queue": "tasks",
		"id":    "live-job",
	})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
