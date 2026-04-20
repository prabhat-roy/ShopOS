package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/session-service/internal/domain"
	"github.com/shopos/session-service/internal/handler"
)

// --------------------------------------------------------------------------
// Mock service
// --------------------------------------------------------------------------

type mockService struct {
	createFn        func(ctx context.Context, req *domain.CreateSessionRequest) (*domain.Session, error)
	validateFn      func(ctx context.Context, id string) (*domain.Session, error)
	getFn           func(ctx context.Context, id string) (*domain.Session, error)
	deleteFn        func(ctx context.Context, id string) error
	listByUserFn    func(ctx context.Context, userID string) ([]*domain.Session, error)
	deleteAllFn     func(ctx context.Context, userID string) error
}

func (m *mockService) Create(ctx context.Context, req *domain.CreateSessionRequest) (*domain.Session, error) {
	return m.createFn(ctx, req)
}
func (m *mockService) Validate(ctx context.Context, id string) (*domain.Session, error) {
	return m.validateFn(ctx, id)
}
func (m *mockService) Get(ctx context.Context, id string) (*domain.Session, error) {
	return m.getFn(ctx, id)
}
func (m *mockService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockService) ListByUser(ctx context.Context, userID string) ([]*domain.Session, error) {
	return m.listByUserFn(ctx, userID)
}
func (m *mockService) DeleteAllByUser(ctx context.Context, userID string) error {
	return m.deleteAllFn(ctx, userID)
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func fakeSession(id, userID string) *domain.Session {
	now := time.Now().UTC()
	return &domain.Session{
		ID:           id,
		UserID:       userID,
		DeviceInfo:   "Chrome/120",
		IPAddress:    "127.0.0.1",
		UserAgent:    "Mozilla/5.0",
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(24 * time.Hour),
	}
}

func newTestServer(svc handler.Servicer) *httptest.Server {
	return httptest.NewServer(handler.New(svc))
}

// --------------------------------------------------------------------------
// GET /healthz
// --------------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	svc := &mockService{}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

// --------------------------------------------------------------------------
// POST /sessions
// --------------------------------------------------------------------------

func TestCreateSession_Returns201(t *testing.T) {
	sess := fakeSession("new-id", "user-1")
	svc := &mockService{
		createFn: func(_ context.Context, req *domain.CreateSessionRequest) (*domain.Session, error) {
			return sess, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	body, _ := json.Marshal(domain.CreateSessionRequest{
		UserID:     "user-1",
		DeviceInfo: "Chrome",
		IPAddress:  "10.0.0.1",
		UserAgent:  "Mozilla/5.0",
	})

	resp, err := http.Post(srv.URL+"/sessions", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /sessions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var got domain.Session
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != "new-id" {
		t.Errorf("expected ID new-id, got %s", got.ID)
	}
}

func TestCreateSession_BadBody_Returns400(t *testing.T) {
	svc := &mockService{}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/sessions", "application/json", bytes.NewReader([]byte("not json")))
	if err != nil {
		t.Fatalf("POST /sessions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --------------------------------------------------------------------------
// GET /sessions/{id}
// --------------------------------------------------------------------------

func TestGetSession_Returns200(t *testing.T) {
	sess := fakeSession("sess-abc", "user-2")
	svc := &mockService{
		validateFn: func(_ context.Context, id string) (*domain.Session, error) {
			if id == "sess-abc" {
				return sess, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sessions/sess-abc")
	if err != nil {
		t.Fatalf("GET /sessions/sess-abc: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var got domain.Session
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != "sess-abc" {
		t.Errorf("expected ID sess-abc, got %s", got.ID)
	}
}

func TestGetSession_NotFound_Returns404(t *testing.T) {
	svc := &mockService{
		validateFn: func(_ context.Context, id string) (*domain.Session, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sessions/ghost")
	if err != nil {
		t.Fatalf("GET /sessions/ghost: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetSession_Expired_Returns401(t *testing.T) {
	svc := &mockService{
		validateFn: func(_ context.Context, id string) (*domain.Session, error) {
			return nil, domain.ErrExpired
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sessions/old-sess")
	if err != nil {
		t.Fatalf("GET /sessions/old-sess: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// --------------------------------------------------------------------------
// DELETE /sessions/{id}
// --------------------------------------------------------------------------

func TestDeleteSession_Returns204(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, id string) error {
			return nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/sessions/sess-del", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /sessions/sess-del: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestDeleteSession_NotFound_Returns404(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, id string) error {
			return domain.ErrNotFound
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/sessions/ghost", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /sessions/ghost: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

// --------------------------------------------------------------------------
// GET /sessions/user/{userID}
// --------------------------------------------------------------------------

func TestListByUser_Returns200(t *testing.T) {
	sessions := []*domain.Session{
		fakeSession("s1", "user-list"),
		fakeSession("s2", "user-list"),
	}
	svc := &mockService{
		listByUserFn: func(_ context.Context, userID string) ([]*domain.Session, error) {
			return sessions, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/sessions/user/user-list")
	if err != nil {
		t.Fatalf("GET /sessions/user/user-list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var got []*domain.Session
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(got))
	}
}

// --------------------------------------------------------------------------
// DELETE /sessions/user/{userID}
// --------------------------------------------------------------------------

func TestDeleteAllByUser_Returns204(t *testing.T) {
	svc := &mockService{
		deleteAllFn: func(_ context.Context, userID string) error {
			return nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/sessions/user/user-all", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /sessions/user/user-all: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}
