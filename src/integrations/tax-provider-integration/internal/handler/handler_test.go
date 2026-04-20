package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/tax-provider-integration/internal/domain"
)

// ---------------------------------------------------------------------------
// mockServicer
// ---------------------------------------------------------------------------

type mockServicer struct {
	calculateFn      func(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error)
	commitFn         func(req domain.CommitRequest) (domain.CommitResponse, error)
	getProviderFn    func(provider domain.TaxProvider) (map[string]interface{}, error)
	validateAddrFn   func(provider domain.TaxProvider, addr domain.TaxAddress) (bool, string, error)
	listProvidersFn  func() []domain.TaxProvider
}

func (m *mockServicer) CalculateTax(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
	return m.calculateFn(req)
}
func (m *mockServicer) CommitTransaction(req domain.CommitRequest) (domain.CommitResponse, error) {
	return m.commitFn(req)
}
func (m *mockServicer) GetProviderInfo(p domain.TaxProvider) (map[string]interface{}, error) {
	return m.getProviderFn(p)
}
func (m *mockServicer) ValidateAddress(p domain.TaxProvider, addr domain.TaxAddress) (bool, string, error) {
	return m.validateAddrFn(p, addr)
}
func (m *mockServicer) ListProviders() []domain.TaxProvider {
	return m.listProvidersFn()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestHandler(m *mockServicer) *Handler {
	return New(m)
}

func do(h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func sampleRequest(provider domain.TaxProvider) domain.TaxCalculationRequest {
	return domain.TaxCalculationRequest{
		Provider:      provider,
		TransactionID: "txn-001",
		FromAddress:   domain.TaxAddress{Street: "1 Merchant St", City: "Seattle", State: "WA", PostalCode: "98101", Country: "US"},
		ToAddress:     domain.TaxAddress{Street: "500 Customer Ave", City: "Portland", State: "OR", PostalCode: "97201", Country: "US"},
		LineItems: []domain.TaxLineItem{
			{ProductID: "p1", SKU: "SKU-001", Description: "Widget", Quantity: 2, UnitPrice: 25.00, Amount: 50.00, TaxCode: "P0000000"},
		},
		Currency: "USD",
	}
}

func sampleResponse(provider domain.TaxProvider) domain.TaxCalculationResponse {
	return domain.TaxCalculationResponse{
		Provider:      provider,
		TransactionID: "txn-001",
		Subtotal:      50.00,
		TotalTax:      0.00,
		Total:         50.00,
		Currency:      "USD",
		Breakdown:     []domain.TaxBreakdownItem{},
		CalculatedAt:  time.Now().UTC(),
	}
}

// ---------------------------------------------------------------------------
// Test 1: Calculate tax via Avalara provider.
// ---------------------------------------------------------------------------
func TestCalculateAvalara(t *testing.T) {
	expected := sampleResponse(domain.ProviderAvalara)
	expected.TotalTax = 3.25
	expected.Total = 53.25

	mock := &mockServicer{
		calculateFn: func(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
			if req.Provider != domain.ProviderAvalara {
				t.Errorf("expected AVALARA provider, got %s", req.Provider)
			}
			return expected, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodPost, "/tax/calculate", sampleRequest(domain.ProviderAvalara))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.TaxCalculationResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Provider != domain.ProviderAvalara {
		t.Errorf("expected AVALARA in response, got %s", resp.Provider)
	}
	if resp.TotalTax != 3.25 {
		t.Errorf("expected totalTax=3.25, got %f", resp.TotalTax)
	}
}

// ---------------------------------------------------------------------------
// Test 2: Calculate tax via TaxJar provider.
// ---------------------------------------------------------------------------
func TestCalculateTaxJar(t *testing.T) {
	expected := sampleResponse(domain.ProviderTaxJar)
	expected.TotalTax = 2.75
	expected.Total = 52.75

	mock := &mockServicer{
		calculateFn: func(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
			return expected, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodPost, "/tax/calculate", sampleRequest(domain.ProviderTaxJar))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.TaxCalculationResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.TotalTax != 2.75 {
		t.Errorf("expected totalTax=2.75, got %f", resp.TotalTax)
	}
}

// ---------------------------------------------------------------------------
// Test 3: Calculate tax via INTERNAL provider.
// ---------------------------------------------------------------------------
func TestCalculateInternal(t *testing.T) {
	expected := sampleResponse(domain.ProviderInternal)
	expected.TotalTax = 0.00

	mock := &mockServicer{
		calculateFn: func(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
			return expected, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodPost, "/tax/calculate", sampleRequest(domain.ProviderInternal))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.TaxCalculationResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Provider != domain.ProviderInternal {
		t.Errorf("expected INTERNAL, got %s", resp.Provider)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Invalid provider returns 422.
// ---------------------------------------------------------------------------
func TestCalculateInvalidProvider(t *testing.T) {
	mock := &mockServicer{
		calculateFn: func(req domain.TaxCalculationRequest) (domain.TaxCalculationResponse, error) {
			return domain.TaxCalculationResponse{}, errors.New("unsupported provider: \"BOGUS\"")
		},
	}
	h := newTestHandler(mock)
	req := sampleRequest("BOGUS")
	rr := do(h, http.MethodPost, "/tax/calculate", req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Test 5: Commit transaction returns 200 with committed=true.
// ---------------------------------------------------------------------------
func TestCommitTransaction(t *testing.T) {
	mock := &mockServicer{
		commitFn: func(req domain.CommitRequest) (domain.CommitResponse, error) {
			return domain.CommitResponse{Committed: true, CommittedAt: time.Now().UTC()}, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodPost, "/tax/commit", domain.CommitRequest{
		TransactionID: "txn-001",
		Provider:      domain.ProviderAvalara,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.CommitResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if !resp.Committed {
		t.Error("expected committed=true")
	}
}

// ---------------------------------------------------------------------------
// Test 6: List providers returns all four providers.
// ---------------------------------------------------------------------------
func TestListProviders(t *testing.T) {
	mock := &mockServicer{
		listProvidersFn: func() []domain.TaxProvider {
			return domain.AllProviders
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodGet, "/tax/providers", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	count, ok := body["count"].(float64)
	if !ok || int(count) != len(domain.AllProviders) {
		t.Errorf("expected count=%d, got %v", len(domain.AllProviders), body["count"])
	}
}

// ---------------------------------------------------------------------------
// Test 7: Get provider info for VERTEX.
// ---------------------------------------------------------------------------
func TestGetProviderInfo(t *testing.T) {
	mock := &mockServicer{
		getProviderFn: func(p domain.TaxProvider) (map[string]interface{}, error) {
			if p != domain.ProviderVertex {
				return nil, errors.New("not found")
			}
			return map[string]interface{}{
				"name":    "Vertex O Series",
				"version": "9.0",
			}, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodGet, "/tax/providers/VERTEX", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["name"] != "Vertex O Series" {
		t.Errorf("unexpected name: %v", body["name"])
	}
}

// ---------------------------------------------------------------------------
// Test 8: Validate address via TaxJar.
// ---------------------------------------------------------------------------
func TestValidateAddress(t *testing.T) {
	mock := &mockServicer{
		validateAddrFn: func(p domain.TaxProvider, addr domain.TaxAddress) (bool, string, error) {
			return true, "PORTLAND, OR 97201, US", nil
		},
	}
	h := newTestHandler(mock)
	body := map[string]interface{}{
		"provider": string(domain.ProviderTaxJar),
		"address": domain.TaxAddress{
			Street:     "500 Customer Ave",
			City:       "Portland",
			State:      "OR",
			PostalCode: "97201",
			Country:    "US",
		},
	}
	rr := do(h, http.MethodPost, "/tax/validate-address", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if valid, _ := resp["valid"].(bool); !valid {
		t.Error("expected valid=true")
	}
	if norm, _ := resp["normalized"].(string); norm == "" {
		t.Error("expected non-empty normalized address")
	}
}
