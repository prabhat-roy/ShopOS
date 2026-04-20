package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/carrier-integration-service/internal/domain"
	"github.com/shopos/carrier-integration-service/internal/handler"
)

// ─── mock registry ────────────────────────────────────────────────────────────

type mockRegistry struct {
	carriers  []*domain.Carrier
	carrier   *domain.Carrier
	carrierErr error
	rates     []domain.RateQuoteResponse
	ratesErr  error
	shipment  *domain.ShipmentResponse
	shipErr   error
	tracking  *domain.TrackResponse
	trackErr  error
}

func (m *mockRegistry) GetCarrier(_ string) (*domain.Carrier, error) {
	return m.carrier, m.carrierErr
}
func (m *mockRegistry) ListCarriers() []*domain.Carrier { return m.carriers }
func (m *mockRegistry) ListActive() []*domain.Carrier   { return m.carriers }
func (m *mockRegistry) GetAllRates(_ domain.RateQuoteRequest) ([]domain.RateQuoteResponse, error) {
	return m.rates, m.ratesErr
}
func (m *mockRegistry) CreateShipment(_ domain.ShipmentRequest) (*domain.ShipmentResponse, error) {
	return m.shipment, m.shipErr
}
func (m *mockRegistry) GetTracking(_ string) (*domain.TrackResponse, error) {
	return m.tracking, m.trackErr
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func newHandler(reg *mockRegistry) http.Handler {
	return handler.NewWithRegistry(reg)
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
	h := newHandler(&mockRegistry{})
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

func TestListCarriers(t *testing.T) {
	reg := &mockRegistry{
		carriers: []*domain.Carrier{
			{ID: "fedex", Name: "FedEx", Code: "FEDEX", Active: true},
			{ID: "ups", Name: "UPS", Code: "UPS", Active: true},
		},
	}
	h := newHandler(reg)
	rr := doRequest(t, h, http.MethodGet, "/carriers", nil)
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

func TestGetCarrier_Found(t *testing.T) {
	reg := &mockRegistry{
		carrier: &domain.Carrier{ID: "fedex", Name: "FedEx", Code: "FEDEX", Active: true},
	}
	h := newHandler(reg)
	rr := doRequest(t, h, http.MethodGet, "/carriers/fedex", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var c domain.Carrier
	if err := json.NewDecoder(rr.Body).Decode(&c); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if c.ID != "fedex" {
		t.Errorf("expected id fedex, got %q", c.ID)
	}
}

func TestGetCarrier_NotFound(t *testing.T) {
	reg := &mockRegistry{carrierErr: domain.ErrCarrierNotFound}
	h := newHandler(reg)
	rr := doRequest(t, h, http.MethodGet, "/carriers/unknown", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetRates_Success(t *testing.T) {
	reg := &mockRegistry{
		rates: []domain.RateQuoteResponse{
			{Carrier: "FedEx", Service: "Ground", EstimatedDays: 5, Price: 12.50, Currency: "USD"},
			{Carrier: "UPS", Service: "2Day", EstimatedDays: 2, Price: 28.75, Currency: "USD"},
		},
	}
	h := newHandler(reg)
	body := domain.RateQuoteRequest{
		FromPostal: "10001", ToPostal: "90210",
		CountryFrom: "US", CountryTo: "US",
		WeightKg: 2.5,
	}
	rr := doRequest(t, h, http.MethodPost, "/carriers/rates", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if total, ok := resp["total"].(float64); !ok || int(total) != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestGetRates_MissingPostal(t *testing.T) {
	h := newHandler(&mockRegistry{})
	body := domain.RateQuoteRequest{WeightKg: 1.0} // missing fromPostal/toPostal
	rr := doRequest(t, h, http.MethodPost, "/carriers/rates", body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateShipment_Success(t *testing.T) {
	reg := &mockRegistry{
		shipment: &domain.ShipmentResponse{
			TrackingNumber: "FEDEX-ABCD1234",
			LabelURL:       "https://labels.shopos.internal/fedex/FEDEX-ABCD1234.pdf",
			Carrier:        "FedEx",
			Service:        "Ground",
			Cost:           12.50,
			Currency:       "USD",
		},
	}
	h := newHandler(reg)
	body := domain.ShipmentRequest{
		Service:  "Ground",
		WeightKg: 2.0,
		FromAddress: domain.Address{
			Name: "Sender", Line1: "123 Main St", City: "New York",
			State: "NY", PostalCode: "10001", Country: "US",
		},
		ToAddress: domain.Address{
			Name: "Receiver", Line1: "456 Elm St", City: "Los Angeles",
			State: "CA", PostalCode: "90210", Country: "US",
		},
	}
	rr := doRequest(t, h, http.MethodPost, "/carriers/fedex/shipments", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.ShipmentResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.TrackingNumber == "" {
		t.Error("expected non-empty tracking number")
	}
}

func TestCreateShipment_CarrierNotFound(t *testing.T) {
	reg := &mockRegistry{shipErr: domain.ErrCarrierNotFound}
	h := newHandler(reg)
	body := domain.ShipmentRequest{Service: "Ground"}
	rr := doRequest(t, h, http.MethodPost, "/carriers/unknown/shipments", body)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetTracking_Success(t *testing.T) {
	reg := &mockRegistry{
		tracking: &domain.TrackResponse{
			TrackingNumber: "FEDEX-ABCD1234",
			Carrier:        "FedEx",
			Status:         "OUT_FOR_DELIVERY",
			Events: []domain.TrackEvent{
				{Timestamp: "2024-01-01T10:00:00Z", Location: "Origin", Description: "Picked up", Status: "PICKED_UP"},
			},
		},
	}
	h := newHandler(reg)
	rr := doRequest(t, h, http.MethodGet, "/carriers/tracking/FEDEX-ABCD1234", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.TrackResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Events) == 0 {
		t.Error("expected at least one tracking event")
	}
}

func TestGetTracking_NotFound(t *testing.T) {
	reg := &mockRegistry{trackErr: domain.ErrTrackingNotFound}
	h := newHandler(reg)
	rr := doRequest(t, h, http.MethodGet, "/carriers/tracking/UNKNOWN-0000", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// TestGetRates_RegistryError ensures a registry error is propagated as 500.
func TestGetRates_RegistryError(t *testing.T) {
	reg := &mockRegistry{ratesErr: errors.New("internal registry error")}
	h := newHandler(reg)
	body := domain.RateQuoteRequest{
		FromPostal: "10001", ToPostal: "90210",
		CountryFrom: "US", CountryTo: "US",
		WeightKg: 1.0,
	}
	rr := doRequest(t, h, http.MethodPost, "/carriers/rates", body)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
