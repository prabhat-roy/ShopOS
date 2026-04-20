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
	"github.com/shopos/credit-service/internal/domain"
	"github.com/shopos/credit-service/internal/handler"
	"github.com/shopos/credit-service/internal/service"
)

// ---- mock -------------------------------------------------------------------

type mockServicer struct {
	createAccount      func(customerID uuid.UUID, limit float64, currency string) (*domain.CreditAccount, error)
	getCreditAccount   func(id uuid.UUID) (*domain.CreditAccount, error)
	getByCustomerID    func(customerID uuid.UUID) (*domain.CreditAccount, error)
	chargeCredit       func(accountID uuid.UUID, amount float64, ref, desc string) (*domain.CreditTransaction, error)
	makePayment        func(accountID uuid.UUID, amount float64, ref string) (*domain.CreditTransaction, error)
	adjustCreditLimit  func(accountID uuid.UUID, newLimit float64) (*domain.CreditAccount, error)
	suspendAccount     func(id uuid.UUID) error
	closeAccount       func(id uuid.UUID) error
	getTransactionHist func(accountID uuid.UUID, limit int) ([]domain.CreditTransaction, error)
}

func (m *mockServicer) CreateCreditAccount(customerID uuid.UUID, limit float64, currency string) (*domain.CreditAccount, error) {
	return m.createAccount(customerID, limit, currency)
}
func (m *mockServicer) GetCreditAccount(id uuid.UUID) (*domain.CreditAccount, error) {
	return m.getCreditAccount(id)
}
func (m *mockServicer) GetByCustomerID(customerID uuid.UUID) (*domain.CreditAccount, error) {
	return m.getByCustomerID(customerID)
}
func (m *mockServicer) ChargeCredit(accountID uuid.UUID, amount float64, ref, desc string) (*domain.CreditTransaction, error) {
	return m.chargeCredit(accountID, amount, ref, desc)
}
func (m *mockServicer) MakePayment(accountID uuid.UUID, amount float64, ref string) (*domain.CreditTransaction, error) {
	return m.makePayment(accountID, amount, ref)
}
func (m *mockServicer) AdjustCreditLimit(accountID uuid.UUID, newLimit float64) (*domain.CreditAccount, error) {
	return m.adjustCreditLimit(accountID, newLimit)
}
func (m *mockServicer) SuspendAccount(id uuid.UUID) error { return m.suspendAccount(id) }
func (m *mockServicer) CloseAccount(id uuid.UUID) error   { return m.closeAccount(id) }
func (m *mockServicer) GetTransactionHistory(accountID uuid.UUID, limit int) ([]domain.CreditTransaction, error) {
	return m.getTransactionHist(accountID, limit)
}

// Verify interface compliance at compile time.
var _ service.Servicer = (*mockServicer)(nil)

// ---- helpers ----------------------------------------------------------------

func newTestServer(svc service.Servicer) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(svc)
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func sampleAccount() *domain.CreditAccount {
	return &domain.CreditAccount{
		ID:              uuid.New(),
		CustomerID:      uuid.New(),
		CreditLimit:     5000,
		AvailableCredit: 5000,
		UsedCredit:      0,
		Currency:        "USD",
		Status:          domain.StatusActive,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

// ---- tests ------------------------------------------------------------------

// Test 1: GET /healthz returns 200 and status ok.
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

// Test 2: POST /credit-accounts creates account and returns 201.
func TestCreateAccount(t *testing.T) {
	acc := sampleAccount()
	mock := &mockServicer{
		createAccount: func(_ uuid.UUID, _ float64, _ string) (*domain.CreditAccount, error) {
			return acc, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	body := map[string]any{
		"customer_id":  acc.CustomerID.String(),
		"credit_limit": 5000.0,
		"currency":     "USD",
	}
	b, _ := json.Marshal(body)
	resp, err := http.Post(srv.URL+"/credit-accounts", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var result domain.CreditAccount
	json.NewDecoder(resp.Body).Decode(&result)
	if result.ID != acc.ID {
		t.Errorf("account ID mismatch")
	}
}

// Test 3: POST /credit-accounts with bad JSON returns 400.
func TestCreateAccount_BadJSON(t *testing.T) {
	srv := newTestServer(&mockServicer{})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/credit-accounts", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 4: GET /credit-accounts/{id} returns account.
func TestGetAccount(t *testing.T) {
	acc := sampleAccount()
	mock := &mockServicer{
		getCreditAccount: func(id uuid.UUID) (*domain.CreditAccount, error) {
			if id == acc.ID {
				return acc, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/credit-accounts/" + acc.ID.String())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 5: GET /credit-accounts/{id} for missing account returns 404.
func TestGetAccount_NotFound(t *testing.T) {
	mock := &mockServicer{
		getCreditAccount: func(_ uuid.UUID) (*domain.CreditAccount, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/credit-accounts/" + uuid.New().String())
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 6: POST /credit-accounts/{id}/charge returns transaction on success.
func TestChargeCredit(t *testing.T) {
	acc := sampleAccount()
	tx := &domain.CreditTransaction{
		ID:        uuid.New(),
		AccountID: acc.ID,
		Type:      domain.TxCharge,
		Amount:    100,
		CreatedAt: time.Now().UTC(),
	}
	mock := &mockServicer{
		chargeCredit: func(_ uuid.UUID, _ float64, _, _ string) (*domain.CreditTransaction, error) {
			return tx, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"amount": 100.0, "reference": "ORD-1"})
	resp, err := http.Post(srv.URL+"/credit-accounts/"+acc.ID.String()+"/charge", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 7: Charge on inactive account returns 422.
func TestChargeCredit_Inactive(t *testing.T) {
	acc := sampleAccount()
	mock := &mockServicer{
		chargeCredit: func(_ uuid.UUID, _ float64, _, _ string) (*domain.CreditTransaction, error) {
			return nil, domain.ErrAccountInactive
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"amount": 100.0})
	resp, err := http.Post(srv.URL+"/credit-accounts/"+acc.ID.String()+"/charge", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
}

// Test 8: POST /credit-accounts/{id}/payment returns 200.
func TestMakePayment(t *testing.T) {
	acc := sampleAccount()
	tx := &domain.CreditTransaction{
		ID:        uuid.New(),
		AccountID: acc.ID,
		Type:      domain.TxPayment,
		Amount:    50,
		CreatedAt: time.Now().UTC(),
	}
	mock := &mockServicer{
		makePayment: func(_ uuid.UUID, _ float64, _ string) (*domain.CreditTransaction, error) {
			return tx, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"amount": 50.0, "reference": "PAY-1"})
	resp, err := http.Post(srv.URL+"/credit-accounts/"+acc.ID.String()+"/payment", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 9: POST /credit-accounts/{id}/suspend returns 204.
func TestSuspendAccount(t *testing.T) {
	acc := sampleAccount()
	mock := &mockServicer{
		suspendAccount: func(_ uuid.UUID) error { return nil },
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/credit-accounts/"+acc.ID.String()+"/suspend", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 10: GET /credit-accounts/{id}/transactions returns list.
func TestListTransactions(t *testing.T) {
	acc := sampleAccount()
	txs := []domain.CreditTransaction{
		{ID: uuid.New(), AccountID: acc.ID, Type: domain.TxCharge, Amount: 200, CreatedAt: time.Now().UTC()},
	}
	mock := &mockServicer{
		getTransactionHist: func(_ uuid.UUID, _ int) ([]domain.CreditTransaction, error) {
			return txs, nil
		},
	}
	srv := newTestServer(mock)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/credit-accounts/" + acc.ID.String() + "/transactions")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string][]domain.CreditTransaction
	json.NewDecoder(resp.Body).Decode(&result)
	if len(result["transactions"]) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(result["transactions"]))
	}
}

// Ensure errors package is used (avoids import error in minimal test builds).
var _ = errors.New
