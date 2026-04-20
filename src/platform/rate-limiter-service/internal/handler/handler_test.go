package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/rate-limiter-service/internal/domain"
	"github.com/shopos/rate-limiter-service/internal/handler"
)

// mockSvc implements handler.Servicer.
type mockSvc struct {
	policy   *domain.Policy
	policies []*domain.Policy
	resp     *domain.CheckResponse
	err      error
}

func (m *mockSvc) CreatePolicy(_ context.Context, _ *domain.CreatePolicyRequest) (*domain.Policy, error) {
	return m.policy, m.err
}
func (m *mockSvc) GetPolicy(_ context.Context, _ string) (*domain.Policy, error) {
	return m.policy, m.err
}
func (m *mockSvc) ListPolicies(_ context.Context) ([]*domain.Policy, error) {
	return m.policies, m.err
}
func (m *mockSvc) UpdatePolicy(_ context.Context, _ string, _ *domain.UpdatePolicyRequest) (*domain.Policy, error) {
	return m.policy, m.err
}
func (m *mockSvc) DeletePolicy(_ context.Context, _ string) error { return m.err }
func (m *mockSvc) Check(_ context.Context, _ domain.CheckRequest) (*domain.CheckResponse, error) {
	return m.resp, m.err
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

func TestListPolicies(t *testing.T) {
	svc := &mockSvc{policies: []*domain.Policy{
		{ID: "1", Key: "api:search", Name: "Search", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/policies", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Policy `json:"items"`
		Total int              `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(body.Items))
	}
}

func TestCreatePolicy(t *testing.T) {
	p := &domain.Policy{ID: "1", Key: "api:login", Name: "Login"}
	h := build(&mockSvc{policy: p})

	b, _ := json.Marshal(domain.CreatePolicyRequest{
		Key: "api:login", Name: "Login", Limit: 10, WindowSecs: 60, Enabled: true,
	})
	req := httptest.NewRequest(http.MethodPost, "/policies", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreatePolicyBadBody(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/policies", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetPolicyNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/policies/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeletePolicy(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/policies/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCheckAllowed(t *testing.T) {
	svc := &mockSvc{resp: &domain.CheckResponse{Allowed: true, Remaining: 9}}
	h := build(svc)

	b, _ := json.Marshal(domain.CheckRequest{PolicyKey: "api:search", Subject: "192.168.1.1", Cost: 1})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.CheckResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Allowed {
		t.Error("expected allowed=true")
	}
}

func TestCheckDenied(t *testing.T) {
	svc := &mockSvc{resp: &domain.CheckResponse{Allowed: false, Remaining: 0, RetryAfter: 60}}
	h := build(svc)

	b, _ := json.Marshal(domain.CheckRequest{PolicyKey: "api:search", Subject: "192.168.1.1", Cost: 1})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestCheckBadBody(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
