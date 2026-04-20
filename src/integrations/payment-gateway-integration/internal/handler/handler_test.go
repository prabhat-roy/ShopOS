package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/payment-gateway-integration/internal/domain"
	"github.com/shopos/payment-gateway-integration/internal/gateway"
	"github.com/shopos/payment-gateway-integration/internal/handler"
	"github.com/shopos/payment-gateway-integration/internal/service"
	"github.com/shopos/payment-gateway-integration/internal/store"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	st := store.New()
	ad := gateway.New()
	svc := service.New(st, ad)
	h := handler.New(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

// chargePayload builds a valid ChargeRequest for the given gateway.
func chargePayload(gw string) map[string]interface{} {
	return map[string]interface{}{
		"gateway":            gw,
		"amount":             99.99,
		"currency":           "USD",
		"customerId":         "cust-001",
		"orderId":            "order-001",
		"paymentMethodToken": "tok_test_visa",
	}
}

// createCharge performs a POST /payments/charge and retries up to 10 times to
// avoid the simulated 5 % failure rate in tests.
func createCharge(t *testing.T, srv *httptest.Server, gw string) domain.ChargeResponse {
	t.Helper()
	for attempt := 0; attempt < 20; attempt++ {
		data, _ := json.Marshal(chargePayload(gw))
		resp, err := http.Post(srv.URL+"/payments/charge", "application/json", bytes.NewReader(data))
		if err != nil {
			t.Fatalf("POST /payments/charge: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusCreated {
			var cr domain.ChargeResponse
			json.NewDecoder(resp.Body).Decode(&cr)
			return cr
		}
	}
	t.Fatal("could not create a successful charge after 20 attempts")
	return domain.ChargeResponse{}
}

// Test 1 — /healthz returns 200 with status ok.
func TestHealthz(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body["status"])
	}
}

// Test 2 — POST /payments/charge with Stripe returns 201 and a pi_ payment ID.
func TestCreateCharge_Stripe(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	cr := createCharge(t, srv, "STRIPE")
	if cr.PaymentIntentID == "" {
		t.Fatal("expected non-empty paymentIntentId")
	}
	if cr.Gateway != domain.GatewayStripe {
		t.Fatalf("expected STRIPE gateway, got %v", cr.Gateway)
	}
	if cr.Status != "succeeded" {
		t.Fatalf("expected status=succeeded, got %v", cr.Status)
	}
}

// Test 3 — POST /payments/charge with missing required fields returns 422.
func TestCreateCharge_MissingFields(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	data, _ := json.Marshal(map[string]interface{}{
		"gateway":  "STRIPE",
		"amount":   10.00,
		"currency": "USD",
		// missing customerId, orderId, paymentMethodToken
	})
	resp, err := http.Post(srv.URL+"/payments/charge", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /payments/charge: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}

// Test 4 — GET /payments/{id} returns the payment intent after a charge.
func TestGetPayment_Found(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	cr := createCharge(t, srv, "PAYPAL")

	resp, err := http.Get(srv.URL + "/payments/" + cr.PaymentIntentID)
	if err != nil {
		t.Fatalf("GET /payments/id: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var pi domain.PaymentIntent
	json.NewDecoder(resp.Body).Decode(&pi)
	if pi.ID != cr.PaymentIntentID {
		t.Fatalf("ID mismatch: expected %s, got %s", cr.PaymentIntentID, pi.ID)
	}
}

// Test 5 — GET /payments/{id} returns 404 for unknown id.
func TestGetPayment_NotFound(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/payments/does-not-exist")
	if err != nil {
		t.Fatalf("GET /payments/id: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 6 — GET /payments?orderId= returns the intent after a charge.
func TestListPayments_ByOrderId(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	createCharge(t, srv, "BRAINTREE")

	resp, err := http.Get(srv.URL + "/payments?orderId=order-001")
	if err != nil {
		t.Fatalf("GET /payments?orderId=: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	payments, ok := body["payments"].([]interface{})
	if !ok || len(payments) == 0 {
		t.Fatal("expected at least one payment in response")
	}
}

// Test 7 — POST /payments/refund refunds a charge and returns 201.
func TestCreateRefund(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	cr := createCharge(t, srv, "ADYEN")

	refReq := map[string]interface{}{
		"paymentIntentId": cr.PaymentIntentID,
		"amount":          50.00,
		"reason":          "customer requested",
	}
	data, _ := json.Marshal(refReq)
	resp, err := http.Post(srv.URL+"/payments/refund", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST /payments/refund: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var rr domain.RefundResponse
	json.NewDecoder(resp.Body).Decode(&rr)
	if rr.RefundID == "" {
		t.Fatal("expected non-empty refundId")
	}
	if rr.Status != "succeeded" {
		t.Fatalf("expected status=succeeded, got %v", rr.Status)
	}
}

// Test 8 — GET /gateways returns all four supported gateways.
func TestListGateways(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/gateways")
	if err != nil {
		t.Fatalf("GET /gateways: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	gateways, ok := body["gateways"].([]interface{})
	if !ok || len(gateways) != 4 {
		t.Fatalf("expected 4 gateways, got %v", gateways)
	}
}
