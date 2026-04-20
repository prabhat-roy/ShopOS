package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/crm-integration-service/internal/adapter"
	"github.com/shopos/crm-integration-service/internal/domain"
	"github.com/shopos/crm-integration-service/internal/handler"
	"github.com/shopos/crm-integration-service/internal/service"
	"github.com/shopos/crm-integration-service/internal/store"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	st := store.New()
	ad := adapter.New()
	svc := service.New(st, ad)
	h := handler.New(svc)
	mux := http.NewServeMux()
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

// Test 2 — POST /crm/contacts/sync with Salesforce contacts returns 202.
func TestSyncContacts_Salesforce_Success(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"crmSystem": "SALESFORCE",
		"customers": []map[string]interface{}{
			{"id": "cust-001", "firstName": "Alice", "lastName": "Smith", "email": "alice@example.com", "phone": "+1-555-0001"},
		},
	}
	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.URL+"/crm/contacts/sync", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST contacts/sync: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	var result domain.SyncResult
	json.NewDecoder(resp.Body).Decode(&result)
	if result.SyncID == "" {
		t.Fatal("expected non-empty syncId")
	}
	if result.ItemsSynced != 1 {
		t.Fatalf("expected 1 item synced, got %d", result.ItemsSynced)
	}
}

// Test 3 — POST /crm/contacts/sync with invalid CRM system returns 400.
func TestSyncContacts_InvalidCrm(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"crmSystem": "UNKNOWN_CRM",
		"customers": []map[string]interface{}{{"id": "x"}},
	}
	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.URL+"/crm/contacts/sync", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST contacts/sync: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 4 — POST /crm/deals/sync with HubSpot orders returns 202.
func TestSyncDeals_HubSpot_Success(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	reqBody := map[string]interface{}{
		"crmSystem": "HUBSPOT",
		"orders": []map[string]interface{}{
			{"id": "order-001", "total": 149.99, "currency": "USD"},
		},
	}
	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(srv.URL+"/crm/deals/sync", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST deals/sync: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	var result domain.SyncResult
	json.NewDecoder(resp.Body).Decode(&result)
	if result.EntityType != "deal" {
		t.Fatalf("expected entityType=deal, got %v", result.EntityType)
	}
}

// Test 5 — GET /crm/contacts returns empty list initially.
func TestListContacts_Empty(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/crm/contacts?crmSystem=ZOHO")
	if err != nil {
		t.Fatalf("GET /crm/contacts: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	count := int(body["count"].(float64))
	if count != 0 {
		t.Fatalf("expected count=0, got %d", count)
	}
}

// Test 6 — GET /crm/contacts/{crmSystem}/{crmId} returns 404 for unknown contact.
func TestGetContact_NotFound(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/crm/contacts/SALESFORCE/nonexistent-id")
	if err != nil {
		t.Fatalf("GET contacts/id: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 7 — GET /crm/sync-history returns results after a sync.
func TestSyncHistory(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	// Create a sync first.
	reqBody := map[string]interface{}{
		"crmSystem": "PIPEDRIVE",
		"customers": []map[string]interface{}{
			{"id": "c1", "firstName": "Bob", "lastName": "Jones", "email": "bob@example.com"},
		},
	}
	data, _ := json.Marshal(reqBody)
	http.Post(srv.URL+"/crm/contacts/sync", "application/json", bytes.NewReader(data))

	resp, err := http.Get(srv.URL + "/crm/sync-history?crmSystem=PIPEDRIVE&limit=10")
	if err != nil {
		t.Fatalf("GET /crm/sync-history: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	count := int(body["count"].(float64))
	if count < 1 {
		t.Fatalf("expected at least 1 history entry, got %d", count)
	}
}

// Test 8 — GET /crm/field-mappings/HUBSPOT returns non-empty mappings.
func TestFieldMappings_HubSpot(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/crm/field-mappings/HUBSPOT")
	if err != nil {
		t.Fatalf("GET field-mappings/HUBSPOT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	if body["crmSystem"] != "HUBSPOT" {
		t.Fatalf("expected crmSystem=HUBSPOT, got %v", body["crmSystem"])
	}
	mappings, ok := body["mappings"].(map[string]interface{})
	if !ok || len(mappings) == 0 {
		t.Fatal("expected non-empty mappings")
	}
}
