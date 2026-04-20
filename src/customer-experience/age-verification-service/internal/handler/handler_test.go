package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/age-verification-service/internal/domain"
	"github.com/shopos/age-verification-service/internal/handler"
	"github.com/shopos/age-verification-service/internal/verifier"
)

func newHandler() http.Handler {
	return handler.New(verifier.New())
}

func doRequest(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func dobYearsAgo(years int) string {
	return time.Now().AddDate(-years, 0, 0).Format("2006-01-02")
}

// Test 1: GET /healthz returns 200 with status ok
func TestHealthz(t *testing.T) {
	h := newHandler()
	rec := doRequest(h, http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok")
	}
}

// Test 2: POST /verify with valid adult returns 200 and verified=true
func TestSingleVerify_ValidAdult(t *testing.T) {
	h := newHandler()
	req := domain.VerificationRequest{
		CustomerID:      "cust-1",
		DateOfBirth:     dobYearsAgo(25),
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	rec := doRequest(h, http.MethodPost, "/verify", req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	var result domain.VerificationResult
	json.NewDecoder(rec.Body).Decode(&result)
	if !result.Verified {
		t.Errorf("expected verified=true")
	}
}

// Test 3: POST /verify with underage customer returns 200 and verified=false
func TestSingleVerify_Underage(t *testing.T) {
	h := newHandler()
	req := domain.VerificationRequest{
		CustomerID:      "cust-2",
		DateOfBirth:     dobYearsAgo(16),
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	rec := doRequest(h, http.MethodPost, "/verify", req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result domain.VerificationResult
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Verified {
		t.Errorf("expected verified=false for 16yo")
	}
}

// Test 4: POST /verify with invalid date returns 400
func TestSingleVerify_InvalidDate(t *testing.T) {
	h := newHandler()
	req := domain.VerificationRequest{
		CustomerID:      "cust-3",
		DateOfBirth:     "not-a-date",
		Country:         "UK",
		ProductCategory: "alcohol",
	}
	rec := doRequest(h, http.MethodPost, "/verify", req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// Test 5: POST /verify/batch returns results for all requests
func TestBatchVerify_ReturnsAllResults(t *testing.T) {
	h := newHandler()
	reqs := []domain.VerificationRequest{
		{CustomerID: "c1", DateOfBirth: dobYearsAgo(25), Country: "UK", ProductCategory: "alcohol"},
		{CustomerID: "c2", DateOfBirth: dobYearsAgo(16), Country: "UK", ProductCategory: "alcohol"},
	}
	rec := doRequest(h, http.MethodPost, "/verify/batch", reqs)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	var results []domain.VerificationResult
	json.NewDecoder(rec.Body).Decode(&results)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Verified {
		t.Errorf("expected result[0] verified=true")
	}
	if results[1].Verified {
		t.Errorf("expected result[1] verified=false")
	}
}

// Test 6: GET /min-age returns correct minimum age
func TestMinAge_ReturnsMinAge(t *testing.T) {
	h := newHandler()
	rec := doRequest(h, http.MethodGet, "/min-age?country=US&category=alcohol", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	minAge, ok := resp["minAge"].(float64)
	if !ok {
		t.Fatalf("minAge not found in response: %v", resp)
	}
	if int(minAge) != 21 {
		t.Errorf("expected minAge=21 for US/alcohol, got %v", minAge)
	}
}
