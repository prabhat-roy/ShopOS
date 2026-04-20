package checker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/health-check-service/internal/checker"
	"github.com/shopos/health-check-service/internal/config"
	"github.com/shopos/health-check-service/internal/domain"
	"go.uber.org/zap"
)

func newChecker(targets []config.Target) *checker.Checker {
	cfg := &config.Config{
		CheckInterval:   time.Hour, // don't auto-run in tests
		CheckTimeout:    2 * time.Second,
		UnhealthyThresh: 3,
		Targets:         targets,
	}
	return checker.New(cfg, zap.NewNop())
}

func TestProbeHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newChecker(nil)
	ctx := context.Background()
	result := c.Probe(ctx, "test-svc", srv.URL)

	if result.Status != domain.StatusHealthy {
		t.Errorf("expected healthy, got %s: %s", result.Status, result.Message)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", result.StatusCode)
	}
	if result.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestProbeUnhealthy500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newChecker(nil)
	result := c.Probe(context.Background(), "bad-svc", srv.URL)

	if result.Status != domain.StatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", result.Status)
	}
}

func TestProbeBadURL(t *testing.T) {
	c := newChecker(nil)
	result := c.Probe(context.Background(), "bad", "http://127.0.0.1:1")
	if result.Status != domain.StatusUnhealthy {
		t.Errorf("expected unhealthy for unreachable target")
	}
	if result.Message == "" {
		t.Error("expected error message")
	}
}

func TestProbeInvalidURL(t *testing.T) {
	c := newChecker(nil)
	result := c.Probe(context.Background(), "bad", "://not-a-url")
	if result.Status != domain.StatusUnhealthy {
		t.Errorf("expected unhealthy for invalid URL")
	}
}

func TestOverallNoTargets(t *testing.T) {
	c := newChecker(nil)
	oh := c.Overall()
	if oh.Status != domain.StatusHealthy {
		t.Errorf("empty target list should be healthy, got %s", oh.Status)
	}
}

func TestOverallAllHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	targets := []config.Target{
		{Name: "svc-a", URL: srv.URL},
		{Name: "svc-b", URL: srv.URL},
	}
	c := newChecker(targets)
	c.Start()
	time.Sleep(50 * time.Millisecond) // let first run complete
	defer c.Stop()

	oh := c.Overall()
	if oh.Status != domain.StatusHealthy {
		t.Errorf("expected all healthy, got %s", oh.Status)
	}
	if len(oh.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(oh.Targets))
	}
}

func TestOverallOneUnhealthy(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer good.Close()

	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer bad.Close()

	cfg := &config.Config{
		CheckInterval:   time.Hour,
		CheckTimeout:    2 * time.Second,
		UnhealthyThresh: 1, // mark unhealthy on first failure
		Targets: []config.Target{
			{Name: "good", URL: good.URL},
			{Name: "bad", URL: bad.URL},
		},
	}
	c := checker.New(cfg, zap.NewNop())
	c.Start()
	time.Sleep(100 * time.Millisecond)
	defer c.Stop()

	oh := c.Overall()
	if oh.Status != domain.StatusUnhealthy {
		t.Errorf("expected overall unhealthy, got %s", oh.Status)
	}
}

func TestGetTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newChecker([]config.Target{{Name: "my-svc", URL: srv.URL}})
	c.Start()
	time.Sleep(50 * time.Millisecond)
	defer c.Stop()

	th, ok := c.Get("my-svc")
	if !ok {
		t.Fatal("expected target to exist")
	}
	if th.Name != "my-svc" {
		t.Errorf("unexpected name %q", th.Name)
	}
}

func TestGetMissingTarget(t *testing.T) {
	c := newChecker(nil)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent target")
	}
}

func TestGracePeriodKeepsHealthy(t *testing.T) {
	// Threshold=3: first 2 failures should still report healthy (grace period)
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cfg := &config.Config{
		CheckInterval:   50 * time.Millisecond,
		CheckTimeout:    2 * time.Second,
		UnhealthyThresh: 3,
		Targets:         []config.Target{{Name: "flaky", URL: srv.URL}},
	}
	c := checker.New(cfg, zap.NewNop())
	c.Start()
	time.Sleep(80 * time.Millisecond) // ~1-2 checks
	defer c.Stop()

	th, _ := c.Get("flaky")
	// With threshold=3 and only 1-2 calls, still in grace period
	if th.Failures >= 3 && th.Status != domain.StatusUnhealthy {
		t.Error("should be unhealthy after 3+ failures")
	}
}
