package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/api-gateway/internal/config"
	"github.com/shopos/api-gateway/internal/router"
)

func testConfig(upstreamURL string) *config.Config {
	return &config.Config{
		Port:            "8080",
		Env:             "development",
		JWTSecret:       "test-secret",
		CORSOrigins:     []string{"*"},
		RateLimitRPS:    1000,
		RateLimitBurst:  1000,
		WebBFFAddr:      upstreamURL,
		MobileBFFAddr:   upstreamURL,
		PartnerBFFAddr:  upstreamURL,
		AdminPortalAddr: upstreamURL,
		UpstreamTimeout: 5 * time.Second,
	}
}

func TestHealthEndpoint_Returns200(t *testing.T) {
	h := router.New(testConfig("http://localhost:9999"))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", body["status"])
	}
}

func TestMetricsEndpoint_Returns200(t *testing.T) {
	h := router.New(testConfig("http://localhost:9999"))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestProtectedRoute_NoToken_Returns401(t *testing.T) {
	h := router.New(testConfig("http://localhost:9999"))
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestPublicAuthRoute_NoToken_ProxiesUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	h := router.New(testConfig(upstream.URL))
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 from upstream, got %d", rr.Code)
	}
}

func TestRequestID_SetOnMissingHeader(t *testing.T) {
	h := router.New(testConfig("http://localhost:9999"))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestCORSOptions_Returns204(t *testing.T) {
	h := router.New(testConfig("http://localhost:9999"))
	req := httptest.NewRequest(http.MethodOptions, "/web/products", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", rr.Code)
	}
}
