package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/webhook-service/internal/domain"
	"github.com/shopos/webhook-service/internal/handler"
)

type mockSvc struct {
	hook  *domain.Webhook
	hooks []*domain.Webhook
	err   error
}

func (m *mockSvc) Create(_ context.Context, _ *domain.CreateWebhookRequest) (*domain.Webhook, error) {
	return m.hook, m.err
}
func (m *mockSvc) Get(_ context.Context, _ string) (*domain.Webhook, error) {
	return m.hook, m.err
}
func (m *mockSvc) List(_ context.Context, _ string) ([]*domain.Webhook, error) {
	return m.hooks, m.err
}
func (m *mockSvc) Update(_ context.Context, _ string, _ *domain.UpdateWebhookRequest) (*domain.Webhook, error) {
	return m.hook, m.err
}
func (m *mockSvc) Delete(_ context.Context, _ string) error {
	return m.err
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

func TestCreateWebhook(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{hook: &domain.Webhook{
		ID: "wh-1", OwnerID: "owner-1", URL: "https://example.com/hook",
		Events: []string{"commerce.order.placed"}, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}}
	h := build(svc)

	body, _ := json.Marshal(domain.CreateWebhookRequest{
		OwnerID: "owner-1",
		URL:     "https://example.com/hook",
		Events:  []string{"commerce.order.placed"},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp domain.Webhook
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ID != "wh-1" {
		t.Errorf("expected id wh-1, got %s", resp.ID)
	}
}

func TestCreateWebhookBadBody(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateWebhookInvalidURL(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrInvalidURL})
	body, _ := json.Marshal(domain.CreateWebhookRequest{URL: "bad-url"})
	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListWebhooks(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{hooks: []*domain.Webhook{
		{ID: "wh-1", OwnerID: "o1", URL: "https://a.com", Events: []string{"x"}, Active: true, CreatedAt: now, UpdatedAt: now},
		{ID: "wh-2", OwnerID: "o1", URL: "https://b.com", Events: []string{"y"}, Active: true, CreatedAt: now, UpdatedAt: now},
	}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhooks?owner_id=o1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["count"].(float64) != 2 {
		t.Errorf("expected count 2, got %v", resp["count"])
	}
}

func TestGetWebhook(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{hook: &domain.Webhook{
		ID: "wh-1", OwnerID: "o1", URL: "https://a.com",
		Events: []string{"commerce.order.placed"}, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/wh-1", nil)
	req.SetPathValue("id", "wh-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetWebhookNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/webhooks/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestUpdateWebhook(t *testing.T) {
	now := time.Now()
	active := true
	svc := &mockSvc{hook: &domain.Webhook{
		ID: "wh-1", OwnerID: "o1", URL: "https://new.com",
		Events: []string{"commerce.order.placed"}, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}}
	h := build(svc)

	newURL := "https://new.com"
	body, _ := json.Marshal(domain.UpdateWebhookRequest{URL: &newURL, Active: &active})
	req := httptest.NewRequest(http.MethodPatch, "/webhooks/wh-1", bytes.NewReader(body))
	req.SetPathValue("id", "wh-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteWebhook(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/webhooks/wh-1", nil)
	req.SetPathValue("id", "wh-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteWebhookNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodDelete, "/webhooks/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
