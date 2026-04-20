package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/partner-bff/internal/config"
	"github.com/shopos/partner-bff/internal/router"
)

func buildHandler(t *testing.T) http.Handler {
	t.Helper()
	cfg := &config.Config{
		Port:             "8083",
		Env:              "development",
		ValidAPIKeys:     map[string]string{"test-key": "partnerA"},
		CatalogAddr:      "localhost:50070",
		InventoryAddr:    "localhost:50074",
		OrderServiceAddr: "localhost:50082",
		WebhookAddr:      "localhost:50055",
		OrgServiceAddr:   "localhost:50160",
	}
	return router.New(cfg)
}

func TestHealthBypassesAPIKey(t *testing.T) {
	h := buildHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMetricsBypassesAPIKey(t *testing.T) {
	h := buildHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUnauthorizedWithoutKey(t *testing.T) {
	h := buildHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/catalog/products", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCatalogRouteWithKey(t *testing.T) {
	h := buildHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/catalog/products", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	// gRPC stub returns 501 Not Implemented
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestOrdersRouteWithKey(t *testing.T) {
	h := buildHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestB2BOrgRouteWithKey(t *testing.T) {
	h := buildHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/b2b/organization", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}
