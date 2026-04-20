package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/api-gateway/internal/middleware"
)

func TestCORS_WildcardOrigin_SetsHeader(t *testing.T) {
	handler := middleware.CORS([]string{"*"})(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected *, got %s", got)
	}
}

func TestCORS_SpecificOrigin_Allowed(t *testing.T) {
	origins := []string{"https://app.shopos.io", "https://admin.shopos.io"}
	handler := middleware.CORS(origins)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.Header.Set("Origin", "https://app.shopos.io")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://app.shopos.io" {
		t.Errorf("expected https://app.shopos.io, got %s", got)
	}
}

func TestCORS_SpecificOrigin_NotAllowed(t *testing.T) {
	origins := []string{"https://app.shopos.io"}
	handler := middleware.CORS(origins)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected empty Allow-Origin for disallowed origin, got %s", got)
	}
}

func TestCORS_OptionsPreflight_Returns204(t *testing.T) {
	handler := middleware.CORS([]string{"*"})(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodOptions, "/web/products", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}
