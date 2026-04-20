package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/push-notification-service/internal/domain"
	"github.com/shopos/push-notification-service/internal/handler"
)

// -------------------------------------------------------------------
// Mock implementations
// -------------------------------------------------------------------

type mockStore struct {
	records map[string]domain.PushRecord
	order   []string
	stats   domain.PushStats
}

func newMockStore() *mockStore {
	return &mockStore{records: make(map[string]domain.PushRecord)}
}

func (m *mockStore) addRecord(r domain.PushRecord) {
	m.records[r.MessageID] = r
	m.order = append(m.order, r.MessageID)
	m.stats.Sent++
	if r.Status == "delivered" {
		m.stats.Delivered++
	} else {
		m.stats.Failed++
	}
}

func (m *mockStore) Get(messageID string) (domain.PushRecord, bool) {
	r, ok := m.records[messageID]
	return r, ok
}

func (m *mockStore) List(limit int) []domain.PushRecord {
	n := len(m.order)
	if limit > n {
		limit = n
	}
	out := make([]domain.PushRecord, 0, limit)
	for i := n - 1; i >= 0 && len(out) < limit; i-- {
		if r, ok := m.records[m.order[i]]; ok {
			out = append(out, r)
		}
	}
	return out
}

func (m *mockStore) Stats() domain.PushStats { return m.stats }

type mockConsumer struct{ running bool }

func (mc *mockConsumer) IsRunning() bool { return mc.running }

// -------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------

func makeRecord(id, status, platform string) domain.PushRecord {
	return domain.PushRecord{
		MessageID:   id,
		DeviceToken: "tok-" + id,
		Platform:    platform,
		Title:       "Test Push",
		Status:      status,
		SentAt:      time.Now().UTC(),
	}
}

func newHandler(st *mockStore, running bool) http.Handler {
	return handler.New(st, &mockConsumer{running: running})
}

// -------------------------------------------------------------------
// Tests
// -------------------------------------------------------------------

func TestHealthz_ReturnsOK(t *testing.T) {
	h := newHandler(newMockStore(), true)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHealthz_ConsumerRunning(t *testing.T) {
	h := newHandler(newMockStore(), true)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["consumer"] != "running" {
		t.Errorf("expected consumer=running, got %q", body["consumer"])
	}
}

func TestHealthz_ConsumerStopped(t *testing.T) {
	h := newHandler(newMockStore(), false)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var body map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["consumer"] != "stopped" {
		t.Errorf("expected consumer=stopped, got %q", body["consumer"])
	}
}

func TestGetPushByID_Found(t *testing.T) {
	st := newMockStore()
	st.addRecord(makeRecord("msg-001", "delivered", "android"))
	h := newHandler(st, true)

	req := httptest.NewRequest(http.MethodGet, "/push/msg-001", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body domain.PushRecord
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.MessageID != "msg-001" {
		t.Errorf("expected messageId=msg-001, got %q", body.MessageID)
	}
}

func TestGetPushByID_NotFound(t *testing.T) {
	h := newHandler(newMockStore(), true)
	req := httptest.NewRequest(http.MethodGet, "/push/nonexistent", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestListPush_ReturnsArray(t *testing.T) {
	st := newMockStore()
	for i := 0; i < 3; i++ {
		st.addRecord(makeRecord("msg-"+string(rune('A'+i)), "delivered", "ios"))
	}
	h := newHandler(st, true)

	req := httptest.NewRequest(http.MethodGet, "/push", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []domain.PushRecord
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if len(body) != 3 {
		t.Errorf("expected 3 records, got %d", len(body))
	}
}

func TestListPush_LimitQueryParam(t *testing.T) {
	st := newMockStore()
	for i := 0; i < 10; i++ {
		st.addRecord(makeRecord(fmt.Sprintf("msg-%02d", i), "delivered", "web"))
	}
	h := newHandler(st, true)

	req := httptest.NewRequest(http.MethodGet, "/push?limit=4", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []domain.PushRecord
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if len(body) != 4 {
		t.Errorf("expected 4 records, got %d", len(body))
	}
}

func TestGetStats(t *testing.T) {
	st := newMockStore()
	st.addRecord(makeRecord("msg-1", "delivered", "ios"))
	st.addRecord(makeRecord("msg-2", "failed", "android"))
	h := newHandler(st, true)

	req := httptest.NewRequest(http.MethodGet, "/push/stats", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body domain.PushStats
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Sent != 2 {
		t.Errorf("expected Sent=2, got %d", body.Sent)
	}
	if body.Delivered != 1 {
		t.Errorf("expected Delivered=1, got %d", body.Delivered)
	}
	if body.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", body.Failed)
	}
}

