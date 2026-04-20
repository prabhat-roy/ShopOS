package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/customs-duties-service/internal/domain"
	"github.com/shopos/customs-duties-service/internal/handler"
)

// ─── mock calculator ──────────────────────────────────────────────────────────

type mockCalc struct {
	dutyResp      *domain.DutyResponse
	dutyErr       error
	hsInfo        *domain.HSCodeInfo
	hsErr         error
	hsCodes       []*domain.HSCodeInfo
	countryRates  *domain.CountryRates
	countryErr    error
}

func (m *mockCalc) Calculate(_ domain.DutyRequest) (*domain.DutyResponse, error) {
	return m.dutyResp, m.dutyErr
}
func (m *mockCalc) GetHSCode(_ string) (*domain.HSCodeInfo, error) {
	return m.hsInfo, m.hsErr
}
func (m *mockCalc) ListHSCodes() []*domain.HSCodeInfo {
	return m.hsCodes
}
func (m *mockCalc) GetCountryRates(_ string) (*domain.CountryRates, error) {
	return m.countryRates, m.countryErr
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newHandler(calc *mockCalc) http.Handler {
	return handler.NewWithCalculator(calc)
}

func doRequest(t *testing.T, h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("failed to encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestHealthz(t *testing.T) {
	h := newHandler(&mockCalc{})
	rr := doRequest(t, h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

// TestCalculate_UStoEU verifies a standard US→DE duty calculation returns
// the expected fields.
func TestCalculate_UStoEU(t *testing.T) {
	calc := &mockCalc{
		dutyResp: &domain.DutyResponse{
			FromCountry:     "US",
			ToCountry:       "DE",
			HSCode:          "8471.30",
			DeclaredValue:   500,
			Currency:        "USD",
			DutyAmount:      0,
			VATAmount:       95,
			TotalLandedCost: 595,
			Breakdown: []domain.DutyLineItem{
				{Description: "VAT — DE", TaxRate: 0.19, TaxAmount: 95},
			},
		},
	}
	h := newHandler(calc)
	body := domain.DutyRequest{
		FromCountry: "US", ToCountry: "DE",
		HSCode: "8471.30", DeclaredValue: 500, Currency: "USD", Quantity: 1,
	}
	rr := doRequest(t, h, http.MethodPost, "/customs/calculate", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.DutyResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ToCountry != "DE" {
		t.Errorf("expected toCountry DE, got %s", resp.ToCountry)
	}
}

// TestCalculate_DeMinimis verifies that a request below de minimis triggers
// the correct response from the calculator.
func TestCalculate_DeMinimis(t *testing.T) {
	calc := &mockCalc{
		dutyResp: &domain.DutyResponse{
			DeMinimisMet: true, DutyAmount: 0, VATAmount: 0,
			TotalLandedCost: 50, Notes: "below de minimis",
		},
	}
	h := newHandler(calc)
	body := domain.DutyRequest{
		FromCountry: "CN", ToCountry: "US",
		HSCode: "9503.00", DeclaredValue: 50, Currency: "USD", Quantity: 1,
	}
	rr := doRequest(t, h, http.MethodPost, "/customs/calculate", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.DutyResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.DeMinimisMet {
		t.Error("expected DeMinimisMet to be true")
	}
	if resp.DutyAmount != 0 {
		t.Errorf("expected 0 duty, got %f", resp.DutyAmount)
	}
}

// TestCalculate_ProhibitedItems verifies prohibited trade pair is flagged.
func TestCalculate_ProhibitedItems(t *testing.T) {
	calc := &mockCalc{
		dutyResp: &domain.DutyResponse{
			ProhibitedItems: true,
			Notes:           "Shipments restricted under trade sanctions.",
		},
	}
	h := newHandler(calc)
	body := domain.DutyRequest{
		FromCountry: "US", ToCountry: "KP",
		HSCode: "8471.30", DeclaredValue: 1000, Currency: "USD", Quantity: 1,
	}
	rr := doRequest(t, h, http.MethodPost, "/customs/calculate", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.DutyResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.ProhibitedItems {
		t.Error("expected ProhibitedItems to be true")
	}
}

// TestCalculate_MissingCountry checks that missing country fields return 400.
func TestCalculate_MissingCountry(t *testing.T) {
	h := newHandler(&mockCalc{})
	body := domain.DutyRequest{DeclaredValue: 100, Quantity: 1}
	rr := doRequest(t, h, http.MethodPost, "/customs/calculate", body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// TestCalculate_InvalidJSON checks that malformed JSON returns 400.
func TestCalculate_InvalidJSON(t *testing.T) {
	h := newHandler(&mockCalc{})
	req := httptest.NewRequest(http.MethodPost, "/customs/calculate", bytes.NewBufferString("{bad json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// TestListHSCodes checks the HS code listing endpoint.
func TestListHSCodes(t *testing.T) {
	calc := &mockCalc{
		hsCodes: []*domain.HSCodeInfo{
			{Code: "8471.30", Description: "Laptops", GeneralRate: 0.00},
			{Code: "6109.10", Description: "T-shirts", GeneralRate: 0.16},
		},
	}
	h := newHandler(calc)
	rr := doRequest(t, h, http.MethodGet, "/customs/hs-codes", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if total, ok := resp["total"].(float64); !ok || int(total) != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

// TestGetHSCode_Found checks lookup of a single HS code.
func TestGetHSCode_Found(t *testing.T) {
	calc := &mockCalc{
		hsInfo: &domain.HSCodeInfo{Code: "8471.30", Description: "Laptops", GeneralRate: 0.00},
	}
	h := newHandler(calc)
	rr := doRequest(t, h, http.MethodGet, "/customs/hs-codes/8471.30", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var info domain.HSCodeInfo
	if err := json.NewDecoder(rr.Body).Decode(&info); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if info.Code != "8471.30" {
		t.Errorf("expected code 8471.30, got %s", info.Code)
	}
}

// TestGetHSCode_NotFound checks 404 for unknown HS code.
func TestGetHSCode_NotFound(t *testing.T) {
	calc := &mockCalc{hsErr: domain.ErrHSCodeNotFound}
	h := newHandler(calc)
	rr := doRequest(t, h, http.MethodGet, "/customs/hs-codes/9999.99", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// TestGetCountryRates_Found checks retrieval of country rates.
func TestGetCountryRates_Found(t *testing.T) {
	calc := &mockCalc{
		countryRates: &domain.CountryRates{
			Country: "DE", VATRate: 0.19, DeMinimiisUSD: 162, Notes: "EU",
		},
	}
	h := newHandler(calc)
	rr := doRequest(t, h, http.MethodGet, "/customs/countries/DE/rates", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var rates domain.CountryRates
	if err := json.NewDecoder(rr.Body).Decode(&rates); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if rates.Country != "DE" {
		t.Errorf("expected country DE, got %s", rates.Country)
	}
}

// TestGetCountryRates_NotFound checks 404 for unknown country.
func TestGetCountryRates_NotFound(t *testing.T) {
	calc := &mockCalc{countryErr: domain.ErrCountryNotFound}
	h := newHandler(calc)
	rr := doRequest(t, h, http.MethodGet, "/customs/countries/ZZ/rates", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// TestCalculate_CalculatorError ensures internal calculator errors return 500.
func TestCalculate_CalculatorError(t *testing.T) {
	calc := &mockCalc{dutyErr: errors.New("unexpected internal error")}
	h := newHandler(calc)
	body := domain.DutyRequest{
		FromCountry: "US", ToCountry: "FR",
		HSCode: "8471.30", DeclaredValue: 200, Currency: "USD", Quantity: 1,
	}
	rr := doRequest(t, h, http.MethodPost, "/customs/calculate", body)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
