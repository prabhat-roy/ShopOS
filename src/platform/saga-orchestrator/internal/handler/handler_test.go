package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/saga-orchestrator/internal/domain"
	"github.com/shopos/saga-orchestrator/internal/handler"
)

type mockSvc struct {
	saga *domain.Saga
	err  error
}

func (m *mockSvc) Start(_ context.Context, _ domain.StartSagaRequest) (*domain.Saga, error) {
	return m.saga, m.err
}
func (m *mockSvc) GetSaga(_ context.Context, _ string) (*domain.Saga, error) {
	return m.saga, m.err
}
func (m *mockSvc) GetSagaByOrder(_ context.Context, _ string) (*domain.Saga, error) {
	return m.saga, m.err
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

func TestStartSaga(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{saga: &domain.Saga{
		ID: "saga-1", OrderID: "order-1",
		State: domain.StateInventoryPending, CreatedAt: now, UpdatedAt: now,
	}}
	h := build(svc)

	b, _ := json.Marshal(domain.StartSagaRequest{
		Type: domain.TypeOrderFulfillment, OrderID: "order-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/sagas", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var body domain.Saga
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.ID != "saga-1" {
		t.Errorf("unexpected ID %q", body.ID)
	}
}

func TestStartSagaBadBody(t *testing.T) {
	h := build(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/sagas", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetSaga(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{saga: &domain.Saga{ID: "saga-1", State: domain.StateCompleted, CreatedAt: now, UpdatedAt: now}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/sagas/saga-1", nil)
	req.SetPathValue("id", "saga-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetSagaNotFound(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/sagas/missing", nil)
	req.SetPathValue("id", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetSagaByOrder(t *testing.T) {
	now := time.Now()
	svc := &mockSvc{saga: &domain.Saga{ID: "saga-1", OrderID: "order-1", CreatedAt: now, UpdatedAt: now}}
	h := build(svc)

	req := httptest.NewRequest(http.MethodGet, "/sagas/order/order-1", nil)
	req.SetPathValue("orderID", "order-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestStartSagaInvalidInput(t *testing.T) {
	h := build(&mockSvc{err: domain.ErrInvalidInput})
	b, _ := json.Marshal(domain.StartSagaRequest{Type: domain.TypeOrderFulfillment})
	req := httptest.NewRequest(http.MethodPost, "/sagas", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
