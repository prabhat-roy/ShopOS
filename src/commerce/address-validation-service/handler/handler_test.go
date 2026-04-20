package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shopos/address-validation-service/domain"
	"github.com/shopos/address-validation-service/handler"
)

func newServer() *httptest.Server {
	h := handler.New()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func postValidate(t *testing.T, srv *httptest.Server, addr domain.Address) domain.ValidationResult {
	t.Helper()
	body, _ := json.Marshal(addr)
	resp, err := http.Post(srv.URL+"/addresses/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("http post: %v", err)
	}
	defer resp.Body.Close()
	var result domain.ValidationResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return result
}

func TestHealthz(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body)
	}
}

func TestValidUSAddress(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	addr := domain.Address{
		Line1:      "123 Main St",
		City:       "New York",
		State:      "NY",
		PostalCode: "10001",
		Country:    "US",
	}
	result := postValidate(t, srv, addr)

	if !result.Valid {
		t.Fatalf("expected valid=true, got issues: %v", result.Issues)
	}
	if result.Confidence <= 0 {
		t.Fatalf("expected confidence > 0, got %f", result.Confidence)
	}
}

func TestInvalidUSPostal(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	addr := domain.Address{
		Line1:      "456 Elm St",
		City:       "Chicago",
		State:      "IL",
		PostalCode: "BADZIP",
		Country:    "US",
	}
	result := postValidate(t, srv, addr)

	if result.Valid {
		t.Fatal("expected valid=false for bad postal code")
	}
	found := false
	for _, issue := range result.Issues {
		if containsAny(issue, "postal", "ZIP") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected postal code issue in %v", result.Issues)
	}
}

func TestMissingCity(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	addr := domain.Address{
		Line1:   "789 Oak Ave",
		Country: "US",
	}
	result := postValidate(t, srv, addr)

	if result.Valid {
		t.Fatal("expected valid=false when city is missing")
	}
	found := false
	for _, issue := range result.Issues {
		if containsAny(issue, "city") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected city issue in %v", result.Issues)
	}
}

func TestValidUKAddress(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	addr := domain.Address{
		Line1:      "10 Downing Street",
		City:       "London",
		PostalCode: "SW1A 2AA",
		Country:    "GB",
	}
	result := postValidate(t, srv, addr)

	if !result.Valid {
		t.Fatalf("expected valid UK address, got issues: %v", result.Issues)
	}
}

func TestValidCanadianAddress(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	addr := domain.Address{
		Line1:      "100 Queen St W",
		City:       "Toronto",
		State:      "ON",
		PostalCode: "M5H 2N2",
		Country:    "CA",
	}
	result := postValidate(t, srv, addr)

	if !result.Valid {
		t.Fatalf("expected valid CA address, got issues: %v", result.Issues)
	}
}

func TestBatchValidation(t *testing.T) {
	srv := newServer()
	defer srv.Close()

	addrs := []domain.Address{
		{Line1: "1 Main St", City: "Boston", State: "MA", PostalCode: "02101", Country: "US"},
		{Line1: "", City: "", Country: ""},
	}
	body, _ := json.Marshal(addrs)
	resp, err := http.Post(srv.URL+"/addresses/validate/batch", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var results []domain.ValidationResult
	json.NewDecoder(resp.Body).Decode(&results)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Valid {
		t.Fatalf("first address should be valid; issues: %v", results[0].Issues)
	}
	if results[1].Valid {
		t.Fatal("second address (empty) should be invalid")
	}
}

// containsAny returns true if s contains any of the given substrings (case-insensitive).
func containsAny(s string, subs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}
