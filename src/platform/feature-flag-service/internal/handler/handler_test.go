package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/feature-flag-service/internal/domain"
	"github.com/shopos/feature-flag-service/internal/handler"
)

// mockSvc implements handler.Servicer.
type mockSvc struct {
	flag    *domain.Flag
	flags   []*domain.Flag
	enabled bool
	err     error
}

func (m *mockSvc) GetFlag(_ context.Context, _ string) (*domain.Flag, error) {
	return m.flag, m.err
}
func (m *mockSvc) ListFlags(_ context.Context) ([]*domain.Flag, error) {
	return m.flags, m.err
}
func (m *mockSvc) CreateFlag(_ context.Context, _ *domain.CreateFlagRequest) (*domain.Flag, error) {
	return m.flag, m.err
}
func (m *mockSvc) UpdateFlag(_ context.Context, _ string, _ *domain.UpdateFlagRequest) (*domain.Flag, error) {
	return m.flag, m.err
}
func (m *mockSvc) DeleteFlag(_ context.Context, _ string) error { return m.err }
func (m *mockSvc) Evaluate(_ context.Context, _ domain.EvalRequest) (bool, error) {
	return m.enabled, m.err
}

var _ handler.Servicer = (*mockSvc)(nil)

func buildMux(svc handler.Servicer) http.Handler {
	mux := http.NewServeMux()
	handler.New(svc).Register(mux)
	return mux
}

func TestHealthEndpoint(t *testing.T) {
	h := buildMux(&mockSvc{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestListFlags(t *testing.T) {
	svc := &mockSvc{flags: []*domain.Flag{
		{ID: "1", Key: "dark-mode", Name: "Dark Mode", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	h := buildMux(svc)

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Flag `json:"items"`
		Total int            `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 {
		t.Errorf("expected 1 flag, got %d", len(body.Items))
	}
}

func TestGetFlag(t *testing.T) {
	svc := &mockSvc{flag: &domain.Flag{ID: "1", Key: "dark-mode", Name: "Dark Mode"}}
	h := buildMux(svc)

	req := httptest.NewRequest(http.MethodGet, "/flags/dark-mode", nil)
	req.SetPathValue("key", "dark-mode")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.Flag
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.Key != "dark-mode" {
		t.Errorf("unexpected key: %q", body.Key)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	svc := &mockSvc{err: domain.ErrNotFound}
	h := buildMux(svc)

	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	req.SetPathValue("key", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateFlag(t *testing.T) {
	created := &domain.Flag{ID: "2", Key: "new-feature", Name: "New Feature"}
	svc := &mockSvc{flag: created}
	h := buildMux(svc)

	payload := domain.CreateFlagRequest{Key: "new-feature", Name: "New Feature", Strategy: domain.StrategyAll}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/flags", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateFlagBadBody(t *testing.T) {
	h := buildMux(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/flags", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateFlagConflict(t *testing.T) {
	svc := &mockSvc{err: domain.ErrAlreadyExists}
	h := buildMux(svc)

	b, _ := json.Marshal(domain.CreateFlagRequest{Key: "dup", Name: "Dup"})
	req := httptest.NewRequest(http.MethodPost, "/flags", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestUpdateFlag(t *testing.T) {
	newName := "Updated"
	updated := &domain.Flag{ID: "1", Key: "feat", Name: newName}
	svc := &mockSvc{flag: updated}
	h := buildMux(svc)

	b, _ := json.Marshal(domain.UpdateFlagRequest{Name: &newName})
	req := httptest.NewRequest(http.MethodPatch, "/flags/1", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDeleteFlag(t *testing.T) {
	h := buildMux(&mockSvc{})
	req := httptest.NewRequest(http.MethodDelete, "/flags/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestEvaluateEnabled(t *testing.T) {
	svc := &mockSvc{enabled: true}
	h := buildMux(svc)

	b, _ := json.Marshal(map[string]string{"Key": "dark-mode", "UserID": "u1"})
	req := httptest.NewRequest(http.MethodPost, "/flags/evaluate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]any
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["enabled"] != true {
		t.Errorf("expected enabled=true, got %v", body["enabled"])
	}
}

func TestEvaluateMissingKey(t *testing.T) {
	h := buildMux(&mockSvc{})
	req := httptest.NewRequest(http.MethodPost, "/flags/evaluate", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
