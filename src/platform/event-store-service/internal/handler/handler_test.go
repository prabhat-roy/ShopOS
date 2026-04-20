package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/event-store-service/internal/domain"
	"github.com/shopos/event-store-service/internal/handler"
)

type mockSvc struct {
	events   []*domain.Event
	event    *domain.Event
	snapshot *domain.Snapshot
	err      error
}

func (m *mockSvc) Append(_ context.Context, _ *domain.AppendRequest) ([]*domain.Event, error) {
	return m.events, m.err
}
func (m *mockSvc) ReadStream(_ context.Context, _ domain.ReadRequest) ([]*domain.Event, error) {
	return m.events, m.err
}
func (m *mockSvc) ReadAll(_ context.Context, _ domain.ReadAllRequest) ([]*domain.Event, error) {
	return m.events, m.err
}
func (m *mockSvc) GetEvent(_ context.Context, _ string) (*domain.Event, error) {
	return m.event, m.err
}
func (m *mockSvc) SaveSnapshot(_ context.Context, _, _ string, _ int64, _ []byte) error {
	return m.err
}
func (m *mockSvc) GetSnapshot(_ context.Context, _ string) (*domain.Snapshot, error) {
	return m.snapshot, m.err
}

var _ handler.Servicer = (*mockSvc)(nil)

func build(svc handler.Servicer) http.Handler {
	mux := http.NewServeMux()
	handler.New(svc).Register(mux)
	return mux
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

func TestAppendEvents(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{events: []*domain.Event{
		{ID: "ev-1", StreamID: "order-1", EventType: "OrderPlaced", Version: 0, OccurredAt: now, RecordedAt: now},
	}}
	h := build(svc)

	body, _ := json.Marshal(domain.AppendRequest{
		StreamType:      "Order",
		ExpectedVersion: -1,
		Events:          []domain.NewEvent{{EventType: "OrderPlaced", Payload: []byte(`{}`)}},
	})
	req := httptest.NewRequest(http.MethodPost, "/streams/order-1/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("streamID", "order-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 1 {
		t.Errorf("expected count 1, got %v", resp["count"])
	}
}

func TestAppendEventsBadBody(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/streams/s1/events", bytes.NewReader([]byte("bad")))
	req.SetPathValue("streamID", "s1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestReadStream(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{events: []*domain.Event{
		{ID: "ev-1", StreamID: "order-1", EventType: "OrderPlaced", Version: 0, OccurredAt: now, RecordedAt: now},
		{ID: "ev-2", StreamID: "order-1", EventType: "OrderPaid", Version: 1, OccurredAt: now, RecordedAt: now},
	}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/streams/order-1/events", nil)
	req.SetPathValue("streamID", "order-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["count"].(float64) != 2 {
		t.Errorf("expected 2 events, got %v", resp["count"])
	}
}

func TestGetEvent(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{event: &domain.Event{ID: "ev-1", StreamID: "s1", EventType: "X", OccurredAt: now, RecordedAt: now}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/events/ev-1", nil)
	req.SetPathValue("id", "ev-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetEventNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/events/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestVersionConflict(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrVersionConflict})
	body, _ := json.Marshal(domain.AppendRequest{
		StreamType: "Order", ExpectedVersion: 5,
		Events: []domain.NewEvent{{EventType: "X", Payload: []byte(`{}`)}},
	})
	req := httptest.NewRequest(http.MethodPost, "/streams/s1/events", bytes.NewReader(body))
	req.SetPathValue("streamID", "s1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestSaveSnapshot(t *testing.T) {
	h := build(&mockSvc{})
	body, _ := json.Marshal(map[string]any{
		"stream_type": "Order",
		"version":     10,
		"state":       []byte(`{"status":"active"}`),
	})
	req := httptest.NewRequest(http.MethodPut, "/streams/order-1/snapshot", bytes.NewReader(body))
	req.SetPathValue("streamID", "order-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestGetSnapshot(t *testing.T) {
	svc := &mockSvc{snapshot: &domain.Snapshot{
		StreamID: "order-1", Version: 10, State: []byte(`{}`), CreatedAt: time.Now(),
	}}
	h := build(svc)
	req := httptest.NewRequest(http.MethodGet, "/streams/order-1/snapshot", nil)
	req.SetPathValue("streamID", "order-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadAll(t *testing.T) {
	h := build(&mockSvc{events: []*domain.Event{}})
	req := httptest.NewRequest(http.MethodGet, "/events?from_seq=0", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
