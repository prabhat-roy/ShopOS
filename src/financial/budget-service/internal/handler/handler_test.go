package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/budget-service/internal/domain"
	"github.com/shopos/budget-service/internal/handler"
	"github.com/shopos/budget-service/internal/service"
)

// ---- mock -------------------------------------------------------------------

type mockServicer struct {
	createBudget      func(dept, name string, period domain.BudgetPeriod, fiscalYear int, start, end time.Time, total float64, currency string) (*domain.Budget, error)
	getBudget         func(id uuid.UUID) (*domain.Budget, error)
	listBudgets       func(dept string, status domain.BudgetStatus, fy int) ([]domain.Budget, error)
	activateBudget    func(id uuid.UUID) error
	closeBudget       func(id uuid.UUID) error
	createAllocation  func(budgetID uuid.UUID, category string, amount float64, notes string) (*domain.BudgetAllocation, error)
	listAllocations   func(budgetID uuid.UUID) ([]domain.BudgetAllocation, error)
	recordSpending    func(budgetID uuid.UUID, allocID *uuid.UUID, category, desc string, amount float64, ref string) (*domain.SpendingRecord, error)
	listSpending      func(budgetID uuid.UUID, limit int) ([]domain.SpendingRecord, error)
	getBudgetSummary  func(id uuid.UUID) (*domain.BudgetSummary, error)
}

func (m *mockServicer) CreateBudget(dept, name string, p domain.BudgetPeriod, fy int, s, e time.Time, t float64, c string) (*domain.Budget, error) {
	return m.createBudget(dept, name, p, fy, s, e, t, c)
}
func (m *mockServicer) GetBudget(id uuid.UUID) (*domain.Budget, error) {
	return m.getBudget(id)
}
func (m *mockServicer) ListBudgets(dept string, status domain.BudgetStatus, fy int) ([]domain.Budget, error) {
	return m.listBudgets(dept, status, fy)
}
func (m *mockServicer) ActivateBudget(id uuid.UUID) error { return m.activateBudget(id) }
func (m *mockServicer) CloseBudget(id uuid.UUID) error    { return m.closeBudget(id) }
func (m *mockServicer) CreateAllocation(bid uuid.UUID, cat string, amt float64, notes string) (*domain.BudgetAllocation, error) {
	return m.createAllocation(bid, cat, amt, notes)
}
func (m *mockServicer) ListAllocations(bid uuid.UUID) ([]domain.BudgetAllocation, error) {
	return m.listAllocations(bid)
}
func (m *mockServicer) RecordSpending(bid uuid.UUID, aid *uuid.UUID, cat, desc string, amt float64, ref string) (*domain.SpendingRecord, error) {
	return m.recordSpending(bid, aid, cat, desc, amt, ref)
}
func (m *mockServicer) ListSpending(bid uuid.UUID, limit int) ([]domain.SpendingRecord, error) {
	return m.listSpending(bid, limit)
}
func (m *mockServicer) GetBudgetSummary(id uuid.UUID) (*domain.BudgetSummary, error) {
	return m.getBudgetSummary(id)
}

// Verify interface compliance.
var _ service.Servicer = (*mockServicer)(nil)

// ---- helpers ----------------------------------------------------------------

func newTestServer(svc service.Servicer) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(svc)
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func sampleBudget() *domain.Budget {
	now := time.Now().UTC()
	return &domain.Budget{
		ID:              uuid.New(),
		Department:      "engineering",
		Name:            "Q1 2026",
		Period:          domain.PeriodQuarterly,
		FiscalYear:      2026,
		StartDate:       now,
		EndDate:         now.AddDate(0, 3, 0),
		TotalAmount:     100000,
		AllocatedAmount: 0,
		SpentAmount:     0,
		RemainingAmount: 100000,
		Currency:        "USD",
		Status:          domain.StatusDraft,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// ---- tests ------------------------------------------------------------------

// Test 1: GET /healthz returns 200 + {"status":"ok"}.
func TestHealthz(t *testing.T) {
	srv := newTestServer(&mockServicer{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: POST /budgets creates a budget and returns 201.
func TestCreateBudget(t *testing.T) {
	b := sampleBudget()
	mock := &mockServicer{
		createBudget: func(_, _ string, _ domain.BudgetPeriod, _ int, _, _ time.Time, _ float64, _ string) (*domain.Budget, error) {
			return b, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	payload := map[string]any{
		"department":   "engineering",
		"name":         "Q1 2026",
		"period":       "QUARTERLY",
		"fiscal_year":  2026,
		"start_date":   b.StartDate.Format(time.RFC3339),
		"end_date":     b.EndDate.Format(time.RFC3339),
		"total_amount": 100000.0,
		"currency":     "USD",
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(srv.URL+"/budgets", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var result domain.Budget
	json.NewDecoder(resp.Body).Decode(&result)
	if result.ID != b.ID {
		t.Errorf("budget ID mismatch")
	}
}

// Test 3: POST /budgets with invalid JSON returns 400.
func TestCreateBudget_BadJSON(t *testing.T) {
	srv := newTestServer(&mockServicer{})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/budgets", "application/json", bytes.NewBufferString("{bad}"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 4: GET /budgets returns list.
func TestListBudgets(t *testing.T) {
	b := sampleBudget()
	mock := &mockServicer{
		listBudgets: func(_ string, _ domain.BudgetStatus, _ int) ([]domain.Budget, error) {
			return []domain.Budget{*b}, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/budgets")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string][]domain.Budget
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result["budgets"]) != 1 {
		t.Errorf("expected 1 budget, got %d", len(result["budgets"]))
	}
}

// Test 5: GET /budgets/{id} returns budget.
func TestGetBudget(t *testing.T) {
	b := sampleBudget()
	mock := &mockServicer{
		getBudget: func(id uuid.UUID) (*domain.Budget, error) {
			if id == b.ID {
				return b, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/budgets/" + b.ID.String())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 6: GET /budgets/{id} for missing budget returns 404.
func TestGetBudget_NotFound(t *testing.T) {
	mock := &mockServicer{
		getBudget: func(_ uuid.UUID) (*domain.Budget, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/budgets/" + uuid.New().String())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 7: PATCH /budgets/{id}/activate returns 204 on success.
func TestActivateBudget(t *testing.T) {
	b := sampleBudget()
	mock := &mockServicer{
		activateBudget: func(_ uuid.UUID) error { return nil },
	}
	srv := newTestServer(mock)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/budgets/"+b.ID.String()+"/activate", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 8: POST /budgets/{id}/allocations returns 201.
func TestCreateAllocation(t *testing.T) {
	b := sampleBudget()
	alloc := &domain.BudgetAllocation{
		ID:              uuid.New(),
		BudgetID:        b.ID,
		Category:        "software",
		AllocatedAmount: 20000,
	}
	mock := &mockServicer{
		createAllocation: func(_ uuid.UUID, _ string, _ float64, _ string) (*domain.BudgetAllocation, error) {
			return alloc, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"category": "software", "amount": 20000.0})
	resp, err := http.Post(srv.URL+"/budgets/"+b.ID.String()+"/allocations", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

// Test 9: POST /budgets/{id}/spending returns 201.
func TestRecordSpending(t *testing.T) {
	b := sampleBudget()
	sr := &domain.SpendingRecord{
		ID:        uuid.New(),
		BudgetID:  b.ID,
		Category:  "software",
		Amount:    500,
		CreatedAt: time.Now().UTC(),
	}
	mock := &mockServicer{
		recordSpending: func(_ uuid.UUID, _ *uuid.UUID, _, _ string, _ float64, _ string) (*domain.SpendingRecord, error) {
			return sr, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"category": "software", "description": "IDE licenses", "amount": 500.0, "reference": "INV-123"})
	resp, err := http.Post(srv.URL+"/budgets/"+b.ID.String()+"/spending", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

// Test 10: GET /budgets/{id}/summary returns utilization.
func TestGetBudgetSummary(t *testing.T) {
	b := sampleBudget()
	b.SpentAmount = 25000
	b.RemainingAmount = 75000
	mock := &mockServicer{
		getBudgetSummary: func(_ uuid.UUID) (*domain.BudgetSummary, error) {
			return &domain.BudgetSummary{
				Budget:      b,
				Allocations: []domain.BudgetAllocation{},
				Utilization: 25.0,
			}, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/budgets/" + b.ID.String() + "/summary")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result domain.BudgetSummary
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Utilization != 25.0 {
		t.Errorf("expected utilization 25.0, got %.1f", result.Utilization)
	}
}

// Ensure errors package is referenced.
var _ = errors.New
