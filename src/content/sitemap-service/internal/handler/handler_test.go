package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopos/sitemap-service/internal/generator"
	"github.com/shopos/sitemap-service/internal/handler"
)

func newHandler() http.Handler {
	gen := generator.New(50000)
	return handler.New(gen, "https://example.com")
}

// Test 1: GET /healthz returns 200 with {"status":"ok"}
func TestHealthz(t *testing.T) {
	h := newHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: POST /sitemaps/generate with empty urls returns 400
func TestGenerateSitemapEmptyURLs(t *testing.T) {
	h := newHandler()

	body := `{"urls":[]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sitemaps/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Test 3: POST /sitemaps/generate with valid URLs returns 200 with XML
func TestGenerateSitemapSuccess(t *testing.T) {
	h := newHandler()

	body := `{"urls":[{"loc":"https://example.com/page1","changefreq":"daily","priority":0.8}]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sitemaps/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/xml") {
		t.Fatalf("expected Content-Type application/xml, got %q", ct)
	}
	output := rec.Body.String()
	if !strings.Contains(output, "<?xml") {
		t.Fatal("expected XML declaration in output")
	}
	if !strings.Contains(output, "https://example.com/page1") {
		t.Fatal("expected URL in sitemap output")
	}
}

// Test 4: POST /sitemaps/index with valid sitemap URLs returns 200 with sitemap index XML
func TestGenerateSitemapIndex(t *testing.T) {
	h := newHandler()

	body := `{"sitemapUrls":["https://example.com/sitemap-1.xml","https://example.com/sitemap-2.xml"]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sitemaps/index", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	output := rec.Body.String()
	if !strings.Contains(output, "<sitemapindex") {
		t.Fatal("expected <sitemapindex> in output")
	}
	if !strings.Contains(output, "sitemap-1.xml") {
		t.Fatal("expected sitemap-1.xml in index output")
	}
}

// Test 5: POST /sitemaps/index with empty sitemapUrls returns 400
func TestGenerateSitemapIndexEmpty(t *testing.T) {
	h := newHandler()

	body := `{"sitemapUrls":[]}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sitemaps/index", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Test 6: GET /robots.txt returns 200 with Sitemap directive pointing to BASE_URL
func TestRobotsTxt(t *testing.T) {
	h := newHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	output := rec.Body.String()
	if !strings.Contains(output, "User-agent:") {
		t.Fatal("missing User-agent directive in robots.txt")
	}
	if !strings.Contains(output, "Sitemap: https://example.com/sitemap.xml") {
		t.Fatalf("expected sitemap URL in robots.txt, got: %s", output)
	}
}

// Test 7: POST /sitemaps/validate splits valid and invalid URLs correctly
func TestValidateURLs(t *testing.T) {
	h := newHandler()

	reqBody := map[string][]string{
		"urls": {
			"https://example.com/valid",
			"not-a-url",
			"http://other.com/also-valid",
			"/relative",
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sitemaps/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string][]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result["valid"]) != 2 {
		t.Fatalf("expected 2 valid URLs, got %d: %v", len(result["valid"]), result["valid"])
	}
	if len(result["invalid"]) != 2 {
		t.Fatalf("expected 2 invalid URLs, got %d: %v", len(result["invalid"]), result["invalid"])
	}
}

// Test 8: Wrong HTTP method on /sitemaps/generate returns 405
func TestGenerateSitemapMethodNotAllowed(t *testing.T) {
	h := newHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/sitemaps/generate", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
