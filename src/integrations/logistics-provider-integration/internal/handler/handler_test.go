package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/logistics-provider-integration/internal/domain"
)

// ---------------------------------------------------------------------------
// mockServicer
// ---------------------------------------------------------------------------

type mockServicer struct {
	createFn        func(req domain.ShipmentRequest) (domain.ShipmentResponse, error)
	getFn           func(trackingNumber string) (domain.ShipmentResponse, error)
	trackFn         func(trackingNumber string, provider domain.Provider) (domain.TrackingResponse, error)
	getRatesFn      func(from, to string, weight float64) ([]domain.RateQuote, error)
	listShipmentsFn func(provider string, limit int) []domain.ShipmentResponse
}

func (m *mockServicer) CreateShipment(req domain.ShipmentRequest) (domain.ShipmentResponse, error) {
	return m.createFn(req)
}
func (m *mockServicer) GetShipment(tn string) (domain.ShipmentResponse, error) {
	return m.getFn(tn)
}
func (m *mockServicer) TrackShipment(tn string, p domain.Provider) (domain.TrackingResponse, error) {
	return m.trackFn(tn, p)
}
func (m *mockServicer) GetRates(from, to string, weight float64) ([]domain.RateQuote, error) {
	return m.getRatesFn(from, to, weight)
}
func (m *mockServicer) ListShipments(provider string, limit int) []domain.ShipmentResponse {
	return m.listShipmentsFn(provider, limit)
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

// ---------------------------------------------------------------------------
// tests
// ---------------------------------------------------------------------------

// Test 1: GET /healthz returns 200 with status ok.
func TestHealthz(t *testing.T) {
	h := newTestHandler(&mockServicer{})
	rr := do(h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: GET /providers returns all providers.
func TestGetProviders(t *testing.T) {
	h := newTestHandler(&mockServicer{})
	rr := do(h, http.MethodGet, "/providers", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	providers, ok := body["providers"]
	if !ok {
		t.Fatal("expected 'providers' key in response")
	}
	list, ok := providers.([]interface{})
	if !ok || len(list) == 0 {
		t.Errorf("expected non-empty providers list")
	}
}

// Test 3: POST /shipments creates a shipment and returns 201.
func TestCreateShipment(t *testing.T) {
	expected := domain.ShipmentResponse{
		TrackingNumber: "1ZTEST000",
		Provider:       domain.ProviderUPS,
		ServiceType:    "UPS_GROUND",
		Cost:           12.50,
		Currency:       "USD",
		CreatedAt:      time.Now().UTC(),
	}
	mock := &mockServicer{
		createFn: func(req domain.ShipmentRequest) (domain.ShipmentResponse, error) {
			return expected, nil
		},
	}
	h := newTestHandler(mock)

	reqBody := domain.ShipmentRequest{
		Provider:    domain.ProviderUPS,
		ServiceType: "UPS_GROUND",
		WeightKg:    2.5,
		FromAddress: domain.Address{Street1: "123 Main St", City: "LA", State: "CA", PostalCode: "90001", Country: "US"},
		ToAddress:   domain.Address{Street1: "456 Elm St", City: "NY", State: "NY", PostalCode: "10001", Country: "US"},
	}
	rr := do(h, http.MethodPost, "/shipments", reqBody)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.ShipmentResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.TrackingNumber != expected.TrackingNumber {
		t.Errorf("expected tracking %q, got %q", expected.TrackingNumber, resp.TrackingNumber)
	}
}

// Test 4: POST /shipments with service error returns 400.
func TestCreateShipment_BadRequest(t *testing.T) {
	mock := &mockServicer{
		createFn: func(req domain.ShipmentRequest) (domain.ShipmentResponse, error) {
			return domain.ShipmentResponse{}, errors.New("unsupported provider: \"INVALID\"")
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodPost, "/shipments", domain.ShipmentRequest{Provider: "INVALID"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Test 5: GET /shipments/{trackingNumber} returns stored shipment.
func TestGetShipment(t *testing.T) {
	expected := domain.ShipmentResponse{TrackingNumber: "7489TEST", Provider: domain.ProviderFedEx}
	mock := &mockServicer{
		getFn: func(tn string) (domain.ShipmentResponse, error) {
			if tn == "7489TEST" {
				return expected, nil
			}
			return domain.ShipmentResponse{}, errors.New("shipment not found: " + tn)
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodGet, "/shipments/7489TEST", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.ShipmentResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.TrackingNumber != "7489TEST" {
		t.Errorf("unexpected tracking number %q", resp.TrackingNumber)
	}
}

// Test 6: GET /shipments/{unknown} returns 404.
func TestGetShipment_NotFound(t *testing.T) {
	mock := &mockServicer{
		getFn: func(tn string) (domain.ShipmentResponse, error) {
			return domain.ShipmentResponse{}, errors.New("shipment not found: " + tn)
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodGet, "/shipments/UNKNOWN", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// Test 7: GET /shipments/{trackingNumber}/track returns tracking info.
func TestTrackShipment(t *testing.T) {
	expected := domain.TrackingResponse{
		TrackingNumber: "JD0TEST",
		Provider:       domain.ProviderDHL,
		Status:         "IN_TRANSIT",
		Events: []domain.TrackingEvent{
			{Location: "Frankfurt Hub", Status: "IN_TRANSIT", Description: "In transit"},
		},
	}
	mock := &mockServicer{
		trackFn: func(tn string, p domain.Provider) (domain.TrackingResponse, error) {
			return expected, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodGet, "/shipments/JD0TEST/track?provider=DHL", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.TrackingResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if resp.Status != "IN_TRANSIT" {
		t.Errorf("expected IN_TRANSIT, got %q", resp.Status)
	}
}

// Test 8: POST /shipments/rates returns rate quotes.
func TestGetRates(t *testing.T) {
	mock := &mockServicer{
		getRatesFn: func(from, to string, weight float64) ([]domain.RateQuote, error) {
			return []domain.RateQuote{
				{Provider: domain.ProviderFedEx, ServiceType: "FEDEX_GROUND", TransitDays: 5, Price: 10.50, Currency: "USD"},
				{Provider: domain.ProviderUPS, ServiceType: "UPS_GROUND", TransitDays: 5, Price: 9.75, Currency: "USD"},
			}, nil
		},
	}
	h := newTestHandler(mock)
	rr := do(h, http.MethodPost, "/shipments/rates", map[string]interface{}{
		"fromPostal": "90001",
		"toPostal":   "10001",
		"weightKg":   2.0,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var body map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	rates, ok := body["rates"].([]interface{})
	if !ok || len(rates) != 2 {
		t.Errorf("expected 2 rate quotes, got %v", body["rates"])
	}
}
