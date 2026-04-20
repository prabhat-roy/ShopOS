package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/dead-letter-service/internal/domain"
	"github.com/shopos/dead-letter-service/internal/handler"
	"log/slog"
	"os"
)

// mockService is a test double for handler.Servicer.
type mockService struct {
	messages map[string]*domain.DeadMessage
	stats    map[string]int64
	saveErr  error
}

func newMockService() *mockService {
	now := time.Now().UTC()
	msgs := map[string]*domain.DeadMessage{
		"msg-1": {
			ID:          "msg-1",
			Topic:       "commerce.order.placed",
			Key:         "order-42",
			Partition:   0,
			Offset:      100,
			Payload:     []byte(`{"order_id":"42"}`),
			ErrorReason: "timeout",
			Status:      domain.StatusPending,
			RetryCount:  0,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	return &mockService{
		messages: msgs,
		stats: map[string]int64{
			"pending":   1,
			"retried":   0,
			"discarded": 0,
		},
	}
}

func (m *mockService) Save(msg *domain.DeadMessage) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.messages[msg.ID] = msg
	return nil
}

func (m *mockService) Get(id string) (*domain.DeadMessage, error) {
	msg, ok := m.messages[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return msg, nil
}

func (m *mockService) List(topic string, status domain.MessageStatus, limit, offset int) ([]*domain.DeadMessage, error) {
	var result []*domain.DeadMessage
	for _, msg := range m.messages {
		if topic != "" && msg.Topic != topic {
			continue
		}
		if status != "" && msg.Status != status {
			continue
		}
		result = append(result, msg)
	}
	return result, nil
}

func (m *mockService) Retry(id string) error {
	msg, ok := m.messages[id]
	if !ok {
		return domain.ErrNotFound
	}
	msg.Status = domain.StatusRetried
	msg.RetryCount++
	return nil
}

func (m *mockService) Discard(id string) error {
	msg, ok := m.messages[id]
	if !ok {
		return domain.ErrNotFound
	}
	msg.Status = domain.StatusDiscarded
	return nil
}

func (m *mockService) Stats() (map[string]int64, error) {
	return m.stats, nil
}

// newTestHandler returns a handler wired to a mock service and a fresh ServeMux.
func newTestHandler(svc handler.Servicer) *http.ServeMux {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	h := handler.New(svc, logger)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// TestHealthz verifies that GET /healthz returns 200 with status ok.
func TestHealthz(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", body["status"])
	}
}

// TestListMessages verifies GET /messages returns an array.
func TestListMessages(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodGet, "/messages", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var msgs []*domain.DeadMessage
	if err := json.NewDecoder(rec.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

// TestListMessagesFilterTopic verifies topic filter returns correct subset.
func TestListMessagesFilterTopic(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodGet, "/messages?topic=nonexistent", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var msgs []*domain.DeadMessage
	if err := json.NewDecoder(rec.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages for unknown topic, got %d", len(msgs))
	}
}

// TestGetMessage verifies GET /messages/{id} returns the correct message.
func TestGetMessage(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodGet, "/messages/msg-1", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var msg domain.DeadMessage
	if err := json.NewDecoder(rec.Body).Decode(&msg); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if msg.ID != "msg-1" {
		t.Fatalf("expected id=msg-1, got %q", msg.ID)
	}
}

// TestGetMessageNotFound verifies GET /messages/{id} returns 404 for missing messages.
func TestGetMessageNotFound(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodGet, "/messages/does-not-exist", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatal("expected error field in response")
	}
}

// TestRetryMessage verifies POST /messages/{id}/retry returns 204 and updates status.
func TestRetryMessage(t *testing.T) {
	svc := newMockService()
	mux := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/messages/msg-1/retry", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if svc.messages["msg-1"].Status != domain.StatusRetried {
		t.Fatalf("expected status=retried, got %q", svc.messages["msg-1"].Status)
	}
	if svc.messages["msg-1"].RetryCount != 1 {
		t.Fatalf("expected retry_count=1, got %d", svc.messages["msg-1"].RetryCount)
	}
}

// TestRetryMessageNotFound verifies POST /messages/{id}/retry returns 404 for missing messages.
func TestRetryMessageNotFound(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodPost, "/messages/ghost/retry", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// TestDiscardMessage verifies POST /messages/{id}/discard returns 204 and updates status.
func TestDiscardMessage(t *testing.T) {
	svc := newMockService()
	mux := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/messages/msg-1/discard", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if svc.messages["msg-1"].Status != domain.StatusDiscarded {
		t.Fatalf("expected status=discarded, got %q", svc.messages["msg-1"].Status)
	}
}

// TestDiscardMessageNotFound verifies POST /messages/{id}/discard returns 404 for missing messages.
func TestDiscardMessageNotFound(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodPost, "/messages/ghost/discard", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// TestStats verifies GET /stats returns the expected shape.
func TestStats(t *testing.T) {
	mux := newTestHandler(newMockService())
	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var counts map[string]int64
	if err := json.NewDecoder(rec.Body).Decode(&counts); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if _, ok := counts["pending"]; !ok {
		t.Fatal("expected pending key in stats")
	}
}

// TestErrNotFoundSentinelIsUsed is a compile-time guard ensuring the sentinel is reachable.
func TestErrNotFoundSentinelIsUsed(t *testing.T) {
	if !errors.Is(domain.ErrNotFound, domain.ErrNotFound) {
		t.Fatal("ErrNotFound sentinel broken")
	}
}
