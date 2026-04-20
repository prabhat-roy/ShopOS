package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shopos/api-gateway/internal/middleware"
)

const testSecret = "test-secret"

func makeToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAuth_PublicPaths_BypassAuth(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	publicPaths := []string{"/healthz", "/metrics", "/auth/login", "/auth/register"}

	for _, path := range publicPaths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("path %s: expected 200, got %d", path, rr.Code)
		}
	}
}

func TestAuth_MissingToken_Returns401(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json Content-Type, got %s", ct)
	}
}

func TestAuth_InvalidToken_Returns401(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":  "user-123",
		"role": "customer",
		"exp":  time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := makeToken(testSecret, claims)

	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_ValidToken_PassesThrough(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":  "user-123",
		"role": "customer",
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	}
	token := makeToken(testSecret, claims)

	var gotUserID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Header.Get("X-User-ID")
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Auth(testSecret)(next)
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected X-User-ID=user-123, got %s", gotUserID)
	}
}

func TestAuth_WrongSecret_Returns401(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	token := makeToken("wrong-secret", claims)

	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))
	req := httptest.NewRequest(http.MethodGet, "/web/products", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
