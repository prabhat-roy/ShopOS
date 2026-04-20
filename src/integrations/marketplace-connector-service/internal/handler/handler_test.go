package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/shopos/marketplace-connector-service/internal/adapter"
	"github.com/shopos/marketplace-connector-service/internal/domain"
	"github.com/shopos/marketplace-connector-service/internal/handler"
	"github.com/shopos/marketplace-connector-service/internal/service"
	"github.com/shopos/marketplace-connector-service/internal/store"
)

func newTestHandler() *handler.Handler {
	st := store.New()
	ad := adapter.New()
	// Use a nil writer; publishEvent will log but not panic in test (writer.WriteMessages will fail gracefully).
	writer := &kafkago.Writer{
		Addr:  kafkago.TCP("localhost:9092"),
		Topic: "test.topic",
	}
	svc := service.New(st, ad, writer, "test.topic")
	return handler.New(svc, ad)
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	h := newTestHandler()
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
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

// Test 2 — POST /marketplace/products/sync with valid Amazon products returns 202.
func TestSyncProducts_Amazon_Success(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"marketplace": "AMAZON",
		"products": []map[string]interface{}{
			{"sku": "SKU-001", "title": "Test Product", "price": 29.99, "quantity": 10, "currency": "USD"},
		},
	}
	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.URL+"/marketplace/products/sync", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST products/sync: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	var rec domain.SyncRecord
	json.NewDecoder(resp.Body).Decode(&rec)
	if rec.ID == "" {
		t.Fatal("expected sync record ID in response")
	}
	if rec.Marketplace != domain.MarketplaceAmazon {
		t.Fatalf("expected AMAZON, got %v", rec.Marketplace)
	}
	if rec.ItemsProcessed != 1 {
		t.Fatalf("expected 1 processed, got %d", rec.ItemsProcessed)
	}
}

// Test 3 — POST /marketplace/products/sync with invalid marketplace returns 400.
func TestSyncProducts_InvalidMarketplace(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"marketplace": "UNKNOWN",
		"products":    []map[string]interface{}{{"sku": "X"}},
	}
	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.URL+"/marketplace/products/sync", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST products/sync: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 4 — POST /marketplace/orders/sync for eBay returns 202.
func TestSyncOrders_Ebay_Success(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"marketplace": "EBAY",
		"limit":       5,
	}
	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.URL+"/marketplace/orders/sync", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST orders/sync: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	var rec domain.SyncRecord
	json.NewDecoder(resp.Body).Decode(&rec)
	if rec.SyncType != domain.SyncTypeOrder {
		t.Fatalf("expected ORDER sync type, got %v", rec.SyncType)
	}
}

// Test 5 — GET /marketplace/syncs/{id} returns 404 for unknown id.
func TestGetSync_NotFound(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/marketplace/syncs/nonexistent-id")
	if err != nil {
		t.Fatalf("GET syncs/id: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 6 — GET /marketplace/syncs/{id} returns the record after a sync.
func TestGetSync_Found(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	// Create a sync first.
	reqBody := map[string]interface{}{
		"marketplace": "ETSY",
		"limit":       2,
	}
	data, _ := json.Marshal(reqBody)
	resp, _ := http.Post(srv.URL+"/marketplace/orders/sync", "application/json", bytes.NewReader(data))
	defer resp.Body.Close()
	var rec domain.SyncRecord
	json.NewDecoder(resp.Body).Decode(&rec)

	// Now retrieve it.
	resp2, err := http.Get(srv.URL + "/marketplace/syncs/" + rec.ID)
	if err != nil {
		t.Fatalf("GET syncs/id: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp2.StatusCode)
	}
}

// Test 7 — GET /marketplace/stats returns marketplace keys.
func TestGetStats(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/marketplace/stats")
	if err != nil {
		t.Fatalf("GET /marketplace/stats: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var stats map[string]int
	json.NewDecoder(resp.Body).Decode(&stats)
	if _, ok := stats["AMAZON"]; !ok {
		t.Fatal("expected AMAZON key in stats")
	}
}

// Test 8 — GET /marketplace/field-mappings/{marketplace} returns mappings for WALMART.
func TestFieldMappings_Walmart(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/marketplace/field-mappings/WALMART")
	if err != nil {
		t.Fatalf("GET field-mappings/WALMART: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["marketplace"] != "WALMART" {
		t.Fatalf("expected marketplace=WALMART, got %v", body["marketplace"])
	}
	mappings, ok := body["mappings"].(map[string]interface{})
	if !ok || len(mappings) == 0 {
		t.Fatal("expected non-empty mappings object")
	}
}
