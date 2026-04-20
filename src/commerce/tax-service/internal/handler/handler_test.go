package handler_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shopos/tax-service/internal/calculator"
	"github.com/shopos/tax-service/internal/domain"
	"github.com/shopos/tax-service/internal/handler"
)

// ─── test setup ──────────────────────────────────────────────────────────────

func setup(t *testing.T) *handler.Handler {
	t.Helper()
	calc := calculator.New()
	logger := log.New(os.Stderr, "[test] ", 0)
	return handler.New(calc, logger)
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("toJSON: %v", err)
	}
	return bytes.NewBuffer(b)
}

func mustDecode(t *testing.T, buf *bytes.Buffer, v any) {
	t.Helper()
	if err := json.NewDecoder(buf).Decode(v); err != nil {
		t.Fatalf("mustDecode: %v", err)
	}
}

// ─── /healthz ────────────────────────────────────────────────────────────────

func TestHealthz_OK(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	mustDecode(t, w.Body, &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ─── POST /tax/calculate ─────────────────────────────────────────────────────

func TestCalculate_USCA_CorrectTax(t *testing.T) {
	h := setup(t)
	req := domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-1", Category: "general", Amount: 100.00, Quantity: 1},
		},
		ShipTo:       domain.Address{Country: "US", State: "CA"},
		Currency:     "USD",
		CustomerType: "b2c",
	}
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	var resp domain.TaxResponse
	mustDecode(t, w.Body, &resp)

	// CA = 7.25%  => $7.25
	if resp.TaxAmount < 7.24 || resp.TaxAmount > 7.26 {
		t.Errorf("US-CA: expected ~7.25, got %.4f", resp.TaxAmount)
	}
	if resp.Subtotal != 100.00 {
		t.Errorf("expected subtotal 100.00, got %.2f", resp.Subtotal)
	}
	if resp.Total < 107.24 || resp.Total > 107.26 {
		t.Errorf("expected total ~107.25, got %.4f", resp.Total)
	}
	if resp.Currency != "USD" {
		t.Errorf("expected currency USD, got %q", resp.Currency)
	}
}

func TestCalculate_EUDE_VAT(t *testing.T) {
	h := setup(t)
	req := domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-2", Amount: 100.00, Quantity: 1},
		},
		ShipTo:       domain.Address{Country: "DE"},
		Currency:     "EUR",
		CustomerType: "b2c",
	}
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	var resp domain.TaxResponse
	mustDecode(t, w.Body, &resp)

	// DE VAT = 19%  => $19.00
	if resp.TaxAmount < 18.99 || resp.TaxAmount > 19.01 {
		t.Errorf("EU-DE: expected ~19.00, got %.4f", resp.TaxAmount)
	}
}

func TestCalculate_EU_B2B_ZeroTax(t *testing.T) {
	h := setup(t)
	req := domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-3", Amount: 500.00, Quantity: 2},
		},
		ShipTo:       domain.Address{Country: "DE"},
		Currency:     "EUR",
		CustomerType: "b2b",
	}
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.TaxResponse
	mustDecode(t, w.Body, &resp)

	if resp.TaxAmount != 0 {
		t.Errorf("EU B2B: expected 0 tax, got %.4f", resp.TaxAmount)
	}
	if resp.TaxRate != 0 {
		t.Errorf("EU B2B: expected 0 rate, got %.4f", resp.TaxRate)
	}
}

func TestCalculate_AU_GST(t *testing.T) {
	h := setup(t)
	req := domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-4", Amount: 200.00, Quantity: 1},
		},
		ShipTo:       domain.Address{Country: "AU"},
		Currency:     "AUD",
		CustomerType: "b2c",
	}
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.TaxResponse
	mustDecode(t, w.Body, &resp)

	// AU GST = 10%  => $20.00
	if resp.TaxAmount < 19.99 || resp.TaxAmount > 20.01 {
		t.Errorf("AU GST: expected ~20.00, got %.4f", resp.TaxAmount)
	}
}

func TestCalculate_MissingCountry_BadRequest(t *testing.T) {
	h := setup(t)
	req := domain.TaxRequest{
		Items: []domain.TaxLineItem{
			{ProductID: "p-1", Amount: 50.00, Quantity: 1},
		},
		// ShipTo.Country intentionally omitted
		ShipTo:       domain.Address{State: "CA"},
		Currency:     "USD",
		CustomerType: "b2c",
	}
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
	var errResp domain.ErrorResponse
	mustDecode(t, w.Body, &errResp)
	if errResp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCalculate_EmptyItems_BadRequest(t *testing.T) {
	h := setup(t)
	req := domain.TaxRequest{
		Items:        []domain.TaxLineItem{},
		ShipTo:       domain.Address{Country: "US", State: "CA"},
		Currency:     "USD",
		CustomerType: "b2c",
	}
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", toJSON(t, req))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCalculate_InvalidJSON_BadRequest(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodPost, "/tax/calculate", bytes.NewBufferString("{invalid"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCalculate_MethodNotAllowed(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/tax/calculate", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ─── GET /tax/rates ──────────────────────────────────────────────────────────

func TestRates_USCA(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/tax/rates?country=US&state=CA", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	var info domain.RateInfo
	mustDecode(t, w.Body, &info)

	if info.Country != "US" {
		t.Errorf("expected country US, got %q", info.Country)
	}
	if info.EffectiveRate < 0.072 || info.EffectiveRate > 0.073 {
		t.Errorf("expected ~0.0725, got %.4f", info.EffectiveRate)
	}
}

func TestRates_MissingCountry_BadRequest(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/tax/rates", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRates_DE(t *testing.T) {
	h := setup(t)
	r := httptest.NewRequest(http.MethodGet, "/tax/rates?country=DE", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var info domain.RateInfo
	mustDecode(t, w.Body, &info)
	if info.EffectiveRate < 0.189 || info.EffectiveRate > 0.191 {
		t.Errorf("expected ~0.19, got %.4f", info.EffectiveRate)
	}
}
