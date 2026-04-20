package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/b2b-credit-limit-service/internal/domain"
	"github.com/shopos/b2b-credit-limit-service/internal/handler"
	"github.com/shopos/b2b-credit-limit-service/internal/service"
)

// --- mock service ---

type mockService struct {
	setLimitFn      func(orgID uuid.UUID, limit float64, currency string) (*domain.OrgCreditLimit, error)
	getLimitFn      func(id uuid.UUID) (*domain.OrgCreditLimit, error)
	getByOrgFn      func(orgID uuid.UUID) (*domain.OrgCreditLimit, error)
	utilizeFn       func(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error)
	paymentFn       func(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error)
	adjustFn        func(orgID uuid.UUID, newLimit float64) (*domain.OrgCreditLimit, error)
	suspendFn       func(orgID uuid.UUID) error
	reviewFn        func(orgID uuid.UUID, score int) (*domain.OrgCreditLimit, error)
	checkFn         func(orgID uuid.UUID, amount float64) (*domain.AvailabilityCheck, error)
	historyFn       func(orgID uuid.UUID, limit int) ([]*domain.CreditTransaction, error)
}

func (m *mockService) SetCreditLimit(orgID uuid.UUID, limit float64, currency string) (*domain.OrgCreditLimit, error) {
	return m.setLimitFn(orgID, limit, currency)
}
func (m *mockService) GetCreditLimit(id uuid.UUID) (*domain.OrgCreditLimit, error) {
	return m.getLimitFn(id)
}
func (m *mockService) GetByOrg(orgID uuid.UUID) (*domain.OrgCreditLimit, error) {
	return m.getByOrgFn(orgID)
}
func (m *mockService) UtilizeCredit(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error) {
	return m.utilizeFn(orgID, amount, ref)
}
func (m *mockService) MakePayment(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error) {
	return m.paymentFn(orgID, amount, ref)
}
func (m *mockService) AdjustLimit(orgID uuid.UUID, newLimit float64) (*domain.OrgCreditLimit, error) {
	return m.adjustFn(orgID, newLimit)
}
func (m *mockService) SuspendOrg(orgID uuid.UUID) error { return m.suspendFn(orgID) }
func (m *mockService) ReviewCredit(orgID uuid.UUID, score int) (*domain.OrgCreditLimit, error) {
	return m.reviewFn(orgID, score)
}
func (m *mockService) CheckAvailability(orgID uuid.UUID, amount float64) (*domain.AvailabilityCheck, error) {
	return m.checkFn(orgID, amount)
}
func (m *mockService) GetCreditHistory(orgID uuid.UUID, limit int) ([]*domain.CreditTransaction, error) {
	return m.historyFn(orgID, limit)
}

// --- helpers ---

func sampleCreditLimit() *domain.OrgCreditLimit {
	return &domain.OrgCreditLimit{
		ID:              uuid.New(),
		OrgID:           uuid.New(),
		CreditLimit:     50000,
		UsedCredit:      10000,
		AvailableCredit: 40000,
		Currency:        "USD",
		Status:          domain.CreditLimitStatusActive,
		RiskScore:       40,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
}

func doRequest(h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
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

// --- tests ---

func TestHealthz(t *testing.T) {
	h := handler.New(&mockService{})
	rr := doRequest(h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestSetCreditLimit_Success(t *testing.T) {
	cl := sampleCreditLimit()
	svc := &mockService{
		setLimitFn: func(orgID uuid.UUID, limit float64, currency string) (*domain.OrgCreditLimit, error) {
			return cl, nil
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{
		"org_id":       cl.OrgID.String(),
		"credit_limit": 50000.0,
		"currency":     "USD",
	}
	rr := doRequest(h, http.MethodPost, "/credit-limits", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGetCreditLimit_Found(t *testing.T) {
	cl := sampleCreditLimit()
	svc := &mockService{
		getLimitFn: func(id uuid.UUID) (*domain.OrgCreditLimit, error) { return cl, nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/credit-limits/"+cl.ID.String(), nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetByOrg_NotFound(t *testing.T) {
	svc := &mockService{
		getByOrgFn: func(orgID uuid.UUID) (*domain.OrgCreditLimit, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/credit-limits/org/"+uuid.New().String(), nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUtilize_Success(t *testing.T) {
	cl := sampleCreditLimit()
	svc := &mockService{
		utilizeFn: func(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error) {
			return cl, nil
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{"amount": 5000.0, "reference": "order-123"}
	rr := doRequest(h, http.MethodPost, "/credit-limits/org/"+cl.OrgID.String()+"/utilize", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUtilize_InsufficientCredit(t *testing.T) {
	svc := &mockService{
		utilizeFn: func(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error) {
			return nil, domain.ErrInsufficientCredit
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{"amount": 999999.0, "reference": "order-big"}
	rr := doRequest(h, http.MethodPost, "/credit-limits/org/"+uuid.New().String()+"/utilize", body)
	if rr.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d", rr.Code)
	}
}

func TestMakePayment_Success(t *testing.T) {
	cl := sampleCreditLimit()
	svc := &mockService{
		paymentFn: func(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error) {
			return cl, nil
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{"amount": 5000.0, "reference": "payment-456"}
	rr := doRequest(h, http.MethodPost, "/credit-limits/org/"+cl.OrgID.String()+"/payment", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSuspend_Success(t *testing.T) {
	svc := &mockService{
		suspendFn: func(orgID uuid.UUID) error { return nil },
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodPost, "/credit-limits/org/"+uuid.New().String()+"/suspend", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestReview_Success(t *testing.T) {
	cl := sampleCreditLimit()
	svc := &mockService{
		reviewFn: func(orgID uuid.UUID, score int) (*domain.OrgCreditLimit, error) { return cl, nil },
	}
	h := handler.New(svc)
	body := map[string]interface{}{"risk_score": 30}
	rr := doRequest(h, http.MethodPost, "/credit-limits/org/"+uuid.New().String()+"/review", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestCheckAvailability_Available(t *testing.T) {
	svc := &mockService{
		checkFn: func(orgID uuid.UUID, amount float64) (*domain.AvailabilityCheck, error) {
			return &domain.AvailabilityCheck{Available: true, AvailableAmount: 40000, RequestedAmount: amount}, nil
		},
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/credit-limits/org/"+uuid.New().String()+"/check?amount=5000", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp domain.AvailabilityCheck
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if !resp.Available {
		t.Error("expected available=true")
	}
}

func TestGetCreditHistory_Success(t *testing.T) {
	orgID := uuid.New()
	svc := &mockService{
		historyFn: func(oID uuid.UUID, limit int) ([]*domain.CreditTransaction, error) {
			return []*domain.CreditTransaction{
				{ID: uuid.New(), OrgID: orgID, Type: domain.TransactionTypeUtilization, Amount: 1000, Balance: 39000, CreatedAt: time.Now().UTC()},
			}, nil
		},
	}
	h := handler.New(svc)
	rr := doRequest(h, http.MethodGet, "/credit-limits/org/"+orgID.String()+"/history", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUtilize_AccountSuspended(t *testing.T) {
	svc := &mockService{
		utilizeFn: func(orgID uuid.UUID, amount float64, ref string) (*domain.OrgCreditLimit, error) {
			return nil, domain.ErrAccountSuspended
		},
	}
	h := handler.New(svc)
	body := map[string]interface{}{"amount": 100.0, "reference": "order-999"}
	rr := doRequest(h, http.MethodPost, "/credit-limits/org/"+uuid.New().String()+"/utilize", body)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
