package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/live-chat-service/internal/domain"
	"github.com/shopos/live-chat-service/internal/handler"
)

// --- Mock service ---

type mockService struct {
	startSessionFn       func(ctx context.Context, customerID string) (domain.ChatSession, error)
	assignAgentFn        func(ctx context.Context, sessionID, agentID string) (domain.ChatSession, error)
	sendMessageFn        func(ctx context.Context, sessionID, senderID, senderType, body string) (domain.ChatMessage, error)
	getSessionFn         func(ctx context.Context, sessionID string) (domain.ChatSession, error)
	getMessagesFn        func(ctx context.Context, sessionID string) ([]domain.ChatMessage, error)
	closeSessionFn       func(ctx context.Context, sessionID string) (domain.ChatSession, error)
	listWaitingSessionFn func(ctx context.Context) ([]domain.ChatSession, error)
}

func (m *mockService) StartSession(ctx context.Context, customerID string) (domain.ChatSession, error) {
	return m.startSessionFn(ctx, customerID)
}
func (m *mockService) AssignAgent(ctx context.Context, sessionID, agentID string) (domain.ChatSession, error) {
	return m.assignAgentFn(ctx, sessionID, agentID)
}
func (m *mockService) SendMessage(ctx context.Context, sessionID, senderID, senderType, body string) (domain.ChatMessage, error) {
	return m.sendMessageFn(ctx, sessionID, senderID, senderType, body)
}
func (m *mockService) GetSession(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	return m.getSessionFn(ctx, sessionID)
}
func (m *mockService) GetMessages(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	return m.getMessagesFn(ctx, sessionID)
}
func (m *mockService) CloseSession(ctx context.Context, sessionID string) (domain.ChatSession, error) {
	return m.closeSessionFn(ctx, sessionID)
}
func (m *mockService) ListWaitingSessions(ctx context.Context) ([]domain.ChatSession, error) {
	return m.listWaitingSessionFn(ctx)
}

// --- Helpers ---

func newSession(id, customerID, status string) domain.ChatSession {
	return domain.ChatSession{
		ID:         id,
		CustomerID: customerID,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func doRequest(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// --- Tests ---

// Test 1: GET /healthz returns 200 with status ok
func TestHealthz(t *testing.T) {
	svc := &mockService{}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp["status"])
	}
}

// Test 2: POST /chat/sessions creates a session and returns 201
func TestCreateSession_Success(t *testing.T) {
	svc := &mockService{
		startSessionFn: func(_ context.Context, customerID string) (domain.ChatSession, error) {
			return newSession("sess-1", customerID, domain.StatusWaiting), nil
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodPost, "/chat/sessions", map[string]string{"customerId": "cust-1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}
	var sess domain.ChatSession
	json.NewDecoder(rec.Body).Decode(&sess)
	if sess.CustomerID != "cust-1" {
		t.Errorf("expected customerId cust-1, got %s", sess.CustomerID)
	}
}

// Test 3: POST /chat/sessions with missing customerId returns 400
func TestCreateSession_MissingCustomerID(t *testing.T) {
	svc := &mockService{}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodPost, "/chat/sessions", map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Test 4: GET /chat/sessions/{id} returns session
func TestGetSession_Found(t *testing.T) {
	svc := &mockService{
		getSessionFn: func(_ context.Context, id string) (domain.ChatSession, error) {
			return newSession(id, "cust-1", domain.StatusActive), nil
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodGet, "/chat/sessions/sess-1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// Test 5: GET /chat/sessions/{id} returns 404 when not found
func TestGetSession_NotFound(t *testing.T) {
	svc := &mockService{
		getSessionFn: func(_ context.Context, id string) (domain.ChatSession, error) {
			return domain.ChatSession{}, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodGet, "/chat/sessions/nope", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// Test 6: POST /chat/sessions/{id}/assign returns 204
func TestAssignAgent_Success(t *testing.T) {
	svc := &mockService{
		assignAgentFn: func(_ context.Context, sessionID, agentID string) (domain.ChatSession, error) {
			return newSession(sessionID, "cust-1", domain.StatusActive), nil
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodPost, "/chat/sessions/sess-1/assign", map[string]string{"agentId": "agent-99"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

// Test 7: POST /chat/sessions/{id}/assign on closed session returns 409
func TestAssignAgent_ClosedSession(t *testing.T) {
	svc := &mockService{
		assignAgentFn: func(_ context.Context, _, _ string) (domain.ChatSession, error) {
			return domain.ChatSession{}, domain.ErrSessionClosed
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodPost, "/chat/sessions/sess-1/assign", map[string]string{"agentId": "a1"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// Test 8: POST /chat/sessions/{id}/messages creates a message and returns 201
func TestSendMessage_Success(t *testing.T) {
	svc := &mockService{
		sendMessageFn: func(_ context.Context, sessionID, senderID, senderType, body string) (domain.ChatMessage, error) {
			return domain.ChatMessage{
				ID:         "msg-1",
				SessionID:  sessionID,
				SenderID:   senderID,
				SenderType: senderType,
				Body:       body,
				SentAt:     time.Now(),
			}, nil
		},
	}
	h := handler.New(svc)
	payload := map[string]string{"senderId": "cust-1", "senderType": "customer", "body": "Hello!"}
	rec := doRequest(h, http.MethodPost, "/chat/sessions/sess-1/messages", payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

// Test 9: POST /chat/sessions/{id}/close returns 204
func TestCloseSession_Success(t *testing.T) {
	svc := &mockService{
		closeSessionFn: func(_ context.Context, id string) (domain.ChatSession, error) {
			return newSession(id, "cust-1", domain.StatusClosed), nil
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodPost, "/chat/sessions/sess-1/close", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

// Test 10: GET /chat/sessions/waiting returns list of waiting sessions
func TestListWaiting(t *testing.T) {
	svc := &mockService{
		listWaitingSessionFn: func(_ context.Context) ([]domain.ChatSession, error) {
			return []domain.ChatSession{
				newSession("s1", "c1", domain.StatusWaiting),
				newSession("s2", "c2", domain.StatusWaiting),
			}, nil
		},
	}
	h := handler.New(svc)
	rec := doRequest(h, http.MethodGet, "/chat/sessions/waiting", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var sessions []domain.ChatSession
	json.NewDecoder(rec.Body).Decode(&sessions)
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

// Ensure errors package is used (compile guard)
var _ = errors.New
