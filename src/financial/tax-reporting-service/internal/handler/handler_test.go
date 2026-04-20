package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/tax-reporting-service/internal/domain"
	"github.com/shopos/tax-reporting-service/internal/handler"
)

// mockServicer is a test double for service.Servicer.
type mockServicer struct {
	recordTaxFn      func(r *domain.TaxRecord) (*domain.TaxRecord, error)
	getTaxRecordFn   func(id string) (*domain.TaxRecord, error)
	listTaxRecordsFn func(f domain.ListFilter) ([]*domain.TaxRecord, error)
	getTaxSummaryFn  func(jurisdiction, period string) ([]*domain.TaxSummary, error)
	generateReportFn func(startDate, endDate, jurisdiction string) ([]*domain.TaxSummary, error)
}

func (m *mockServicer) RecordTax(r *domain.TaxRecord) (*domain.TaxRecord, error) {
	return m.recordTaxFn(r)
}
func (m *mockServicer) GetTaxRecord(id string) (*domain.TaxRecord, error) {
	return m.getTaxRecordFn(id)
}
func (m *mockServicer) ListTaxRecords(f domain.ListFilter) ([]*domain.TaxRecord, error) {
	return m.listTaxRecordsFn(f)
}
func (m *mockServicer) GetTaxSummary(jurisdiction, period string) ([]*domain.TaxSummary, error) {
	return m.getTaxSummaryFn(jurisdiction, period)
}
func (m *mockServicer) GenerateReport(startDate, endDate, jurisdiction string) ([]*domain.TaxSummary, error) {
	return m.generateReportFn(startDate, endDate, jurisdiction)
}

func newFixedRecord() *domain.TaxRecord {
	return &domain.TaxRecord{
		ID:              "rec-001",
		OrderID:         "ord-001",
		CustomerID:      "cust-001",
		Jurisdiction:    "US-CA",
		TaxType:         domain.TaxTypeSalesTax,
		TaxableAmount:   100.0,
		TaxRate:         8.5,
		TaxAmount:       8.5,
		Currency:        "USD",
		TransactionDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Test 1: GET /healthz returns 200 and status ok.
func TestHealthz(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: POST /tax-records returns 201 with created record.
func TestCreateTaxRecord_Success(t *testing.T) {
	rec := newFixedRecord()
	svc := &mockServicer{
		recordTaxFn: func(r *domain.TaxRecord) (*domain.TaxRecord, error) {
			return rec, nil
		},
	}
	h := handler.New(svc)

	body, _ := json.Marshal(rec)
	req := httptest.NewRequest(http.MethodPost, "/tax-records", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var got domain.TaxRecord
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != "rec-001" {
		t.Errorf("expected id rec-001, got %s", got.ID)
	}
}

// Test 3: POST /tax-records with bad JSON returns 400.
func TestCreateTaxRecord_BadJSON(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodPost, "/tax-records", bytes.NewBufferString("{bad json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Test 4: POST /tax-records with service validation error returns 422.
func TestCreateTaxRecord_ValidationError(t *testing.T) {
	svc := &mockServicer{
		recordTaxFn: func(r *domain.TaxRecord) (*domain.TaxRecord, error) {
			return nil, domain.ErrNotFound // simulate a service error
		},
	}
	h := handler.New(svc)
	body, _ := json.Marshal(map[string]string{"orderId": ""})
	req := httptest.NewRequest(http.MethodPost, "/tax-records", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

// Test 5: GET /tax-records/{id} returns record.
func TestGetTaxRecord_Found(t *testing.T) {
	rec := newFixedRecord()
	svc := &mockServicer{
		getTaxRecordFn: func(id string) (*domain.TaxRecord, error) {
			if id == "rec-001" {
				return rec, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodGet, "/tax-records/rec-001", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

// Test 6: GET /tax-records/{id} for missing record returns 404.
func TestGetTaxRecord_NotFound(t *testing.T) {
	svc := &mockServicer{
		getTaxRecordFn: func(id string) (*domain.TaxRecord, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodGet, "/tax-records/missing", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// Test 7: GET /tax-records returns list.
func TestListTaxRecords(t *testing.T) {
	records := []*domain.TaxRecord{newFixedRecord()}
	svc := &mockServicer{
		listTaxRecordsFn: func(f domain.ListFilter) ([]*domain.TaxRecord, error) {
			return records, nil
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodGet, "/tax-records?jurisdiction=US-CA", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []*domain.TaxRecord
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 record, got %d", len(got))
	}
}

// Test 8: POST /tax-reports/generate returns summaries.
func TestGenerateReport_Success(t *testing.T) {
	summaries := []*domain.TaxSummary{
		{Jurisdiction: "US-CA", TaxType: domain.TaxTypeSalesTax, Period: "2024-06 to 2024-06", TotalTaxable: 100, TotalTax: 8.5, TransactionCount: 1},
	}
	svc := &mockServicer{
		generateReportFn: func(startDate, endDate, jurisdiction string) ([]*domain.TaxSummary, error) {
			return summaries, nil
		},
	}
	h := handler.New(svc)
	body, _ := json.Marshal(domain.GenerateReportRequest{StartDate: "2024-06-01", EndDate: "2024-06-30", Jurisdiction: "US-CA"})
	req := httptest.NewRequest(http.MethodPost, "/tax-reports/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []*domain.TaxSummary
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 summary, got %d", len(got))
	}
	if got[0].TotalTax != 8.5 {
		t.Errorf("expected totalTax=8.5, got %f", got[0].TotalTax)
	}
}
