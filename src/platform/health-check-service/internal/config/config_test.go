package config_test

import (
	"testing"

	"github.com/shopos/health-check-service/internal/config"
)

func TestParseTargets(t *testing.T) {
	t.Setenv("HEALTH_TARGETS", "order-service=http://order:8080/healthz,payment-service=http://payment:8080/healthz")
	cfg := config.Load()
	if len(cfg.Targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(cfg.Targets))
	}
	if cfg.Targets[0].Name != "order-service" {
		t.Errorf("expected order-service, got %q", cfg.Targets[0].Name)
	}
	if cfg.Targets[1].URL != "http://payment:8080/healthz" {
		t.Errorf("unexpected URL: %q", cfg.Targets[1].URL)
	}
}

func TestParseTargetsEmpty(t *testing.T) {
	t.Setenv("HEALTH_TARGETS", "")
	cfg := config.Load()
	if len(cfg.Targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(cfg.Targets))
	}
}

func TestDefaultPort(t *testing.T) {
	cfg := config.Load()
	if cfg.Port == "" {
		t.Error("expected non-empty port")
	}
}
