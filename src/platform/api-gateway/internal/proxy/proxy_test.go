package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/api-gateway/internal/proxy"
	"go.uber.org/zap"
)

func TestProxy_ForwardsRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, "", 5*time.Second, zap.NewNop())
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestProxy_StripsPrefixBeforeForwarding(t *testing.T) {
	var receivedPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, "/web", 5*time.Second, zap.NewNop())
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if receivedPath != "/products" {
		t.Errorf("expected upstream to receive /products, got %s", receivedPath)
	}
}

func TestProxy_StripsPrefixToRoot(t *testing.T) {
	var receivedPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, "/web", 5*time.Second, zap.NewNop())
	req := httptest.NewRequest(http.MethodGet, "/web", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if receivedPath != "/" {
		t.Errorf("expected upstream to receive /, got %s", receivedPath)
	}
}

func TestProxy_UpstreamDown_Returns502(t *testing.T) {
	h := proxy.New("http://localhost:19999", "", 1*time.Second, zap.NewNop())
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}

func TestProxy_SetsForwardedHostHeader(t *testing.T) {
	var receivedHost string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHost = r.Header.Get("X-Forwarded-Host")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	h := proxy.New(upstream.URL, "", 5*time.Second, zap.NewNop())
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Host = "api.shopos.io"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if receivedHost != "api.shopos.io" {
		t.Errorf("expected X-Forwarded-Host=api.shopos.io, got %s", receivedHost)
	}
}

func TestProxy_WebSocketDetection(t *testing.T) {
	// WebSocket requests to a non-WS upstream should get connection refused, not a proxy error
	h := proxy.New("http://localhost:19999", "", 1*time.Second, zap.NewNop())
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// WebSocket to unavailable upstream returns 502
	if rr.Code != http.StatusBadGateway && rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 502 or 500 for WS to unavailable upstream, got %d", rr.Code)
	}
}
