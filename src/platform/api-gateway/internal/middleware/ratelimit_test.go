package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/api-gateway/internal/middleware"
	"go.uber.org/zap"
)

func TestRateLimiter_WithinLimit_Passes(t *testing.T) {
	rl := middleware.NewRateLimiter(100, 10, zap.NewNop())
	handler := rl.Middleware(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRateLimiter_ExceedBurst_Returns429(t *testing.T) {
	// burst=2 means only 2 requests allowed immediately
	rl := middleware.NewRateLimiter(0.001, 2, zap.NewNop())
	handler := rl.Middleware(http.HandlerFunc(okHandler))

	var lastCode int
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
		req.RemoteAddr = "10.0.0.1:9999"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		lastCode = rr.Code
	}

	if lastCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 after burst exceeded, got %d", lastCode)
	}
}

func TestRateLimiter_DifferentIPs_IndependentLimits(t *testing.T) {
	rl := middleware.NewRateLimiter(0.001, 1, zap.NewNop())
	handler := rl.Middleware(http.HandlerFunc(okHandler))

	for _, ip := range []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"} {
		req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
		req.RemoteAddr = ip + ":1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("ip %s: expected 200 on first request, got %d", ip, rr.Code)
		}
	}
}
