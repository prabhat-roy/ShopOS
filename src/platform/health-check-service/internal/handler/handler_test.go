package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/health-check-service/internal/domain"
	"github.com/shopos/health-check-service/internal/handler"
)

// mockSource implements handler.HealthSource.
type mockSource struct {
	overall domain.OverallHealth
	targets map[string]domain.TargetHealth
}

func (m *mockSource) Overall() domain.OverallHealth { return m.overall }
func (m *mockSource) Get(name string) (domain.TargetHealth, bool) {
	th, ok := m.targets[name]
	return th, ok
}

var _ handler.HealthSource = (*mockSource)(nil)

func buildMux(src handler.HealthSource) http.Handler {
	mux := http.NewServeMux()
	handler.New(src).Register(mux)
	return mux
}

func TestSelfHealth(t *testing.T) {
	h := buildMux(&mockSource{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected ok, got %q", body["status"])
	}
}

func TestOverallHealthy(t *testing.T) {
	src := &mockSource{
		overall: domain.OverallHealth{
			Status:    domain.StatusHealthy,
			Targets:   map[string]domain.TargetHealth{},
			CheckedAt: time.Now(),
		},
	}
	h := buildMux(src)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.OverallHealth
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Status != domain.StatusHealthy {
		t.Errorf("expected healthy, got %q", body.Status)
	}
}

func TestOverallUnhealthy(t *testing.T) {
	src := &mockSource{
		overall: domain.OverallHealth{
			Status:    domain.StatusUnhealthy,
			Targets:   map[string]domain.TargetHealth{},
			CheckedAt: time.Now(),
		},
	}
	h := buildMux(src)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestSingleTargetHealthy(t *testing.T) {
	src := &mockSource{
		targets: map[string]domain.TargetHealth{
			"order-service": {Name: "order-service", Status: domain.StatusHealthy},
		},
	}
	h := buildMux(src)
	req := httptest.NewRequest(http.MethodGet, "/health/order-service", nil)
	req.SetPathValue("name", "order-service")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.TargetHealth
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Status != domain.StatusHealthy {
		t.Errorf("expected healthy, got %q", body.Status)
	}
}

func TestSingleTargetUnhealthy(t *testing.T) {
	src := &mockSource{
		targets: map[string]domain.TargetHealth{
			"bad-svc": {Name: "bad-svc", Status: domain.StatusUnhealthy, Message: "connection refused"},
		},
	}
	h := buildMux(src)
	req := httptest.NewRequest(http.MethodGet, "/health/bad-svc", nil)
	req.SetPathValue("name", "bad-svc")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestSingleTargetNotFound(t *testing.T) {
	src := &mockSource{targets: map[string]domain.TargetHealth{}}
	h := buildMux(src)
	req := httptest.NewRequest(http.MethodGet, "/health/missing", nil)
	req.SetPathValue("name", "missing")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestOverallWithMultipleTargets(t *testing.T) {
	src := &mockSource{
		overall: domain.OverallHealth{
			Status: domain.StatusHealthy,
			Targets: map[string]domain.TargetHealth{
				"svc-a": {Name: "svc-a", Status: domain.StatusHealthy},
				"svc-b": {Name: "svc-b", Status: domain.StatusHealthy},
			},
			CheckedAt: time.Now(),
		},
	}
	h := buildMux(src)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.OverallHealth
	json.Unmarshal(w.Body.Bytes(), &body)
	if len(body.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(body.Targets))
	}
}
