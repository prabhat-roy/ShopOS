package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/partner-bff/internal/middleware"
	"go.uber.org/zap"
)

var validKeys = map[string]string{"secret-key-1": "partner-A", "secret-key-2": "partner-B"}

func okHandler(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }

func TestAPIKey_HealthBypasses(t *testing.T) {
	h := middleware.APIKey(validKeys, zap.NewNop())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAPIKey_MissingKey_Returns401(t *testing.T) {
	h := middleware.APIKey(validKeys, zap.NewNop())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/catalog/products", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKey_InvalidKey_Returns401(t *testing.T) {
	h := middleware.APIKey(validKeys, zap.NewNop())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/catalog/products", nil)
	req.Header.Set("X-API-Key", "bad-key")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKey_ValidKey_InjectsPartnerID(t *testing.T) {
	var gotPartnerID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPartnerID = r.Header.Get("X-Partner-ID")
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.APIKey(validKeys, zap.NewNop())(next)
	req := httptest.NewRequest(http.MethodGet, "/catalog/products", nil)
	req.Header.Set("X-API-Key", "secret-key-1")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if gotPartnerID != "partner-A" {
		t.Errorf("expected X-Partner-ID=partner-A, got %s", gotPartnerID)
	}
}

func TestAPIKey_ContentTypeOnError(t *testing.T) {
	h := middleware.APIKey(validKeys, zap.NewNop())(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}
