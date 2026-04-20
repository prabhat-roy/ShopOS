package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/expense-management-service/internal/domain"
	"github.com/shopos/expense-management-service/internal/handler"
)

// mockServicer is a test double for service.Servicer.
type mockServicer struct {
	createExpenseFn   func(e *domain.Expense) (*domain.Expense, error)
	getExpenseFn      func(id uuid.UUID) (*domain.Expense, error)
	listExpensesFn    func(f domain.ListFilter) ([]*domain.Expense, error)
	submitExpenseFn   func(id uuid.UUID) (*domain.Expense, error)
	approveExpenseFn  func(id uuid.UUID, approverID uuid.UUID) (*domain.Expense, error)
	rejectExpenseFn   func(id uuid.UUID, reason string) (*domain.Expense, error)
	reimburseExpenseFn func(id uuid.UUID) (*domain.Expense, error)
	deleteExpenseFn   func(id uuid.UUID) error
}

func (m *mockServicer) CreateExpense(e *domain.Expense) (*domain.Expense, error) {
	return m.createExpenseFn(e)
}
func (m *mockServicer) GetExpense(id uuid.UUID) (*domain.Expense, error) {
	return m.getExpenseFn(id)
}
func (m *mockServicer) ListExpenses(f domain.ListFilter) ([]*domain.Expense, error) {
	return m.listExpensesFn(f)
}
func (m *mockServicer) SubmitExpense(id uuid.UUID) (*domain.Expense, error) {
	return m.submitExpenseFn(id)
}
func (m *mockServicer) ApproveExpense(id uuid.UUID, approverID uuid.UUID) (*domain.Expense, error) {
	return m.approveExpenseFn(id, approverID)
}
func (m *mockServicer) RejectExpense(id uuid.UUID, reason string) (*domain.Expense, error) {
	return m.rejectExpenseFn(id, reason)
}
func (m *mockServicer) ReimburseExpense(id uuid.UUID) (*domain.Expense, error) {
	return m.reimburseExpenseFn(id)
}
func (m *mockServicer) DeleteExpense(id uuid.UUID) error {
	return m.deleteExpenseFn(id)
}

var (
	fixedID       = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	fixedEmployee = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	fixedApprover = uuid.MustParse("33333333-3333-3333-3333-333333333333")
)

func newFixedExpense(status domain.ExpenseStatus) *domain.Expense {
	return &domain.Expense{
		ID:          fixedID,
		EmployeeID:  fixedEmployee,
		Category:    domain.CategoryTravel,
		Amount:      150.00,
		Currency:    "USD",
		Description: "Flight to conference",
		ReceiptURL:  "https://receipts.example.com/r1",
		Status:      status,
		CreatedAt:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Test 1: GET /healthz returns 200 with status ok.
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

// Test 2: POST /expenses creates expense with 201.
func TestCreateExpense_Success(t *testing.T) {
	expense := newFixedExpense(domain.StatusDraft)
	svc := &mockServicer{
		createExpenseFn: func(e *domain.Expense) (*domain.Expense, error) {
			return expense, nil
		},
	}
	h := handler.New(svc)

	payload, _ := json.Marshal(expense)
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var got domain.Expense
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != fixedID {
		t.Errorf("expected id %s, got %s", fixedID, got.ID)
	}
	if got.Status != domain.StatusDraft {
		t.Errorf("expected status DRAFT, got %s", got.Status)
	}
}

// Test 3: POST /expenses with bad JSON returns 400.
func TestCreateExpense_BadJSON(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewBufferString("{not json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Test 4: GET /expenses/{id} returns the expense.
func TestGetExpense_Found(t *testing.T) {
	expense := newFixedExpense(domain.StatusDraft)
	svc := &mockServicer{
		getExpenseFn: func(id uuid.UUID) (*domain.Expense, error) {
			if id == fixedID {
				return expense, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodGet, "/expenses/"+fixedID.String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

// Test 5: GET /expenses/{id} for missing id returns 404.
func TestGetExpense_NotFound(t *testing.T) {
	svc := &mockServicer{
		getExpenseFn: func(id uuid.UUID) (*domain.Expense, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodGet, "/expenses/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// Test 6: GET /expenses returns list.
func TestListExpenses(t *testing.T) {
	expenses := []*domain.Expense{newFixedExpense(domain.StatusDraft)}
	svc := &mockServicer{
		listExpensesFn: func(f domain.ListFilter) ([]*domain.Expense, error) {
			return expenses, nil
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodGet, "/expenses?status=DRAFT", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []*domain.Expense
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 expense, got %d", len(got))
	}
}

// Test 7: POST /expenses/{id}/submit returns 204.
func TestSubmitExpense(t *testing.T) {
	expense := newFixedExpense(domain.StatusSubmitted)
	svc := &mockServicer{
		submitExpenseFn: func(id uuid.UUID) (*domain.Expense, error) {
			return expense, nil
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodPost, "/expenses/"+fixedID.String()+"/submit", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

// Test 8: POST /expenses/{id}/approve returns 204.
func TestApproveExpense(t *testing.T) {
	expense := newFixedExpense(domain.StatusApproved)
	svc := &mockServicer{
		approveExpenseFn: func(id uuid.UUID, approverID uuid.UUID) (*domain.Expense, error) {
			return expense, nil
		},
	}
	h := handler.New(svc)
	body, _ := json.Marshal(map[string]string{"approverId": fixedApprover.String()})
	req := httptest.NewRequest(http.MethodPost, "/expenses/"+fixedID.String()+"/approve", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

// Test 9: POST /expenses/{id}/reject returns 204.
func TestRejectExpense(t *testing.T) {
	expense := newFixedExpense(domain.StatusRejected)
	svc := &mockServicer{
		rejectExpenseFn: func(id uuid.UUID, reason string) (*domain.Expense, error) {
			return expense, nil
		},
	}
	h := handler.New(svc)
	body, _ := json.Marshal(map[string]string{"reason": "Receipt missing"})
	req := httptest.NewRequest(http.MethodPost, "/expenses/"+fixedID.String()+"/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

// Test 10: DELETE /expenses/{id} on DRAFT returns 204.
func TestDeleteExpense_Draft(t *testing.T) {
	svc := &mockServicer{
		deleteExpenseFn: func(id uuid.UUID) error {
			return nil
		},
	}
	h := handler.New(svc)
	req := httptest.NewRequest(http.MethodDelete, "/expenses/"+fixedID.String(), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}
