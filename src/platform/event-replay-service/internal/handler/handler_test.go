package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/event-replay-service/internal/domain"
	"github.com/shopos/event-replay-service/internal/handler"
)

// ---- mock -------------------------------------------------------------------

type mockService struct {
	jobs   map[string]*domain.ReplayJob
	nextID string
}

func newMock() *mockService {
	return &mockService{
		jobs:   make(map[string]*domain.ReplayJob),
		nextID: "job-1",
	}
}

func (m *mockService) seedJob(status domain.ReplayStatus) *domain.ReplayJob {
	j := &domain.ReplayJob{
		ID:        m.nextID,
		StreamID:  "stream-1",
		Target:    domain.TargetHTTP,
		Status:    status,
		CreatedAt: time.Now().UTC(),
	}
	m.jobs[j.ID] = j
	return j
}

func (m *mockService) CreateJob(_ context.Context, req *domain.CreateReplayRequest) (*domain.ReplayJob, error) {
	j := &domain.ReplayJob{
		ID:          m.nextID,
		StreamID:    req.StreamID,
		StreamType:  req.StreamType,
		EventType:   req.EventType,
		FromSeq:     req.FromSeq,
		ToSeq:       req.ToSeq,
		Target:      req.Target,
		TargetTopic: req.TargetTopic,
		Status:      domain.StatusPending,
		CreatedAt:   time.Now().UTC(),
	}
	if j.Target == "" {
		j.Target = domain.TargetHTTP
	}
	m.jobs[j.ID] = j
	return j, nil
}

func (m *mockService) GetJob(_ context.Context, id string) (*domain.ReplayJob, error) {
	j, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return j, nil
}

func (m *mockService) ListJobs(_ context.Context) ([]*domain.ReplayJob, error) {
	out := make([]*domain.ReplayJob, 0, len(m.jobs))
	for _, j := range m.jobs {
		out = append(out, j)
	}
	return out, nil
}

func (m *mockService) CancelJob(_ context.Context, id string) error {
	if _, ok := m.jobs[id]; !ok {
		return domain.ErrNotFound
	}
	m.jobs[id].Status = domain.StatusCancelled
	return nil
}

func (m *mockService) RunJob(_ context.Context, id string) error {
	if _, ok := m.jobs[id]; !ok {
		return domain.ErrNotFound
	}
	m.jobs[id].Status = domain.StatusRunning
	return nil
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := handler.New(newMock())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestCreateReplay_Returns202(t *testing.T) {
	h := handler.New(newMock())

	payload := domain.CreateReplayRequest{
		StreamID:   "stream-abc",
		StreamType: "order",
		EventType:  "order.placed",
		FromSeq:    1,
		ToSeq:      100,
		Target:     domain.TargetHTTP,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/replays", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	var job domain.ReplayJob
	if err := json.NewDecoder(w.Body).Decode(&job); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if job.Status != domain.StatusPending {
		t.Errorf("expected status=pending, got %q", job.Status)
	}
}

func TestCreateReplay_BadBody_Returns400(t *testing.T) {
	h := handler.New(newMock())
	req := httptest.NewRequest(http.MethodPost, "/replays", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListReplays_EmptyReturnsArray(t *testing.T) {
	h := handler.New(newMock())
	req := httptest.NewRequest(http.MethodGet, "/replays", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var jobs []*domain.ReplayJob
	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if jobs == nil {
		t.Error("expected non-nil slice")
	}
}

func TestListReplays_ReturnsExistingJobs(t *testing.T) {
	m := newMock()
	m.seedJob(domain.StatusPending)
	h := handler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/replays", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var jobs []*domain.ReplayJob
	if err := json.NewDecoder(w.Body).Decode(&jobs); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(jobs))
	}
}

func TestGetReplay_Found(t *testing.T) {
	m := newMock()
	seeded := m.seedJob(domain.StatusPending)
	h := handler.New(m)

	req := httptest.NewRequest(http.MethodGet, "/replays/"+seeded.ID, nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var job domain.ReplayJob
	if err := json.NewDecoder(w.Body).Decode(&job); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if job.ID != seeded.ID {
		t.Errorf("expected ID=%s, got %s", seeded.ID, job.ID)
	}
}

func TestGetReplay_NotFound_Returns404(t *testing.T) {
	h := handler.New(newMock())
	req := httptest.NewRequest(http.MethodGet, "/replays/does-not-exist", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRunReplay_Returns202(t *testing.T) {
	m := newMock()
	seeded := m.seedJob(domain.StatusPending)
	h := handler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/replays/"+seeded.ID+"/run", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
}

func TestRunReplay_NotFound_Returns404(t *testing.T) {
	h := handler.New(newMock())
	req := httptest.NewRequest(http.MethodPost, "/replays/ghost-id/run", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCancelReplay_Returns204(t *testing.T) {
	m := newMock()
	seeded := m.seedJob(domain.StatusPending)
	h := handler.New(m)

	req := httptest.NewRequest(http.MethodPost, "/replays/"+seeded.ID+"/cancel", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCancelReplay_NotFound_Returns404(t *testing.T) {
	h := handler.New(newMock())
	req := httptest.NewRequest(http.MethodPost, "/replays/ghost-id/cancel", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
