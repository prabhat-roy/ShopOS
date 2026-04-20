package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/api-gateway/internal/middleware"
	"go.uber.org/zap"
)

func TestRecovery_PanicReturns500(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	handler := middleware.Recovery(zap.NewNop())(panicHandler)
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestRecovery_NoPanic_PassesThrough(t *testing.T) {
	handler := middleware.Recovery(zap.NewNop())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
