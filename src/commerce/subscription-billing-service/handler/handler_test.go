package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/subscription-billing-service/domain"
	"github.com/shopos/subscription-billing-service/handler"
	"github.com/shopos/subscription-billing-service/service"
)

// ---- mock service --------------------------------------------------------

type mockSvc struct {
	subscribeFn         func(req service.SubscribeRequest) (*domain.Subscription, error)
	getSubscriptionFn   func(id string) (*domain.Subscription, error)
	listSubscriptionsFn func(customerID string) ([]*domain.Subscription, error)
	cancelFn            func(id string) error
	pauseFn             func(id string) error
	resumeFn            func(id string) error
	listBillingFn       func(subID string) ([]*domain.BillingRecord, error)
}

func (m *mockSvc) Subscribe(req service.SubscribeRequest) (*domain.Subscription, error) {
	return m.subscribeFn(req)
}
func (m *mockSvc) GetSubscription(id string) (*domain.Subscription, error) {
	return m.getSubscriptionFn(id)
}
func (m *mockSvc) ListSubscriptions(customerID string) ([]*domain.Subscription, error) {
	return m.listSubscriptionsFn(customerID)
}
func (m *mockSvc) Cancel(id string) error { return m.cancelFn(id) }
func (m *mockSvc) Pause(id string) error  { return m.pauseFn(id) }
func (m *mockSvc) Resume(id string) error { return m.resumeFn(id) }
func (m *mockSvc) ListBillingRecords(subID string) ([]*domain.BillingRecord, error) {
	return m.listBillingFn(subID)
}

// ---- helpers -------------------------------------------------------------

func newHandler(svc handler.Svc) http.Handler {
	h := handler.New(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

func buildSub(id string) *domain.Subscription {
	return &domain.Subscription{
		ID:            id,
		CustomerID:    "cust-001",
		PlanID:        "plan-monthly",
		ProductID:     "prod-001",
		Status:        domain.SubActive,
		Cycle:         domain.CycleMonthly,
		Price:         9.99,
		Currency:      "USD",
		NextBillingAt: time.Now().UTC().AddDate(0, 1, 0),
		StartedAt:     time.Now().UTC(),
		CreatedAt:     time.Now().UTC(),
	}
}

// ---- GET /healthz --------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := newHandler(&mockSvc{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body)
	}
}

// ---- POST /subscriptions -------------------------------------------------

func TestCreateSubscription_Success(t *testing.T) {
	sub := buildSub("sub-001")
	svc := &mockSvc{
		subscribeFn: func(req service.SubscribeRequest) (*domain.Subscription, error) {
			if req.CustomerID != "cust-001" {
				t.Errorf("unexpected customer_id: %s", req.CustomerID)
			}
			return sub, nil
		},
	}

	body := `{"customer_id":"cust-001","plan_id":"plan-monthly","cycle":"monthly","price":9.99,"currency":"USD"}`
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["id"] != "sub-001" {
		t.Errorf("unexpected id: %v", resp["id"])
	}
}

func TestCreateSubscription_MissingFields(t *testing.T) {
	svc := &mockSvc{}
	body := `{"plan_id":"plan-monthly"}` // missing customer_id
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateSubscription_ServiceError(t *testing.T) {
	svc := &mockSvc{
		subscribeFn: func(_ service.SubscribeRequest) (*domain.Subscription, error) {
			return nil, errors.New("db error")
		},
	}
	body := `{"customer_id":"cust-001","plan_id":"plan-monthly"}`
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}

// ---- GET /subscriptions/{id} ---------------------------------------------

func TestGetSubscription_Found(t *testing.T) {
	sub := buildSub("sub-123")
	svc := &mockSvc{
		getSubscriptionFn: func(id string) (*domain.Subscription, error) {
			if id != "sub-123" {
				t.Errorf("unexpected id: %s", id)
			}
			return sub, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/sub-123", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetSubscription_NotFound(t *testing.T) {
	svc := &mockSvc{
		getSubscriptionFn: func(_ string) (*domain.Subscription, error) {
			return nil, domain.ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/missing-id", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---- GET /subscriptions?customer_id= -------------------------------------

func TestListSubscriptions(t *testing.T) {
	subs := []*domain.Subscription{buildSub("s1"), buildSub("s2")}
	svc := &mockSvc{
		listSubscriptionsFn: func(customerID string) ([]*domain.Subscription, error) {
			if customerID != "cust-001" {
				t.Errorf("unexpected customerID: %s", customerID)
			}
			return subs, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/subscriptions?customer_id=cust-001", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(resp))
	}
}

func TestListSubscriptions_MissingCustomerID(t *testing.T) {
	svc := &mockSvc{}
	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// ---- POST /subscriptions/{id}/cancel -------------------------------------

func TestCancelSubscription_Success(t *testing.T) {
	svc := &mockSvc{
		cancelFn: func(id string) error {
			if id != "sub-abc" {
				t.Errorf("unexpected id: %s", id)
			}
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/sub-abc/cancel", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestCancelSubscription_NotFound(t *testing.T) {
	svc := &mockSvc{
		cancelFn: func(_ string) error { return domain.ErrNotFound },
	}

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/ghost/cancel", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---- POST /subscriptions/{id}/pause --------------------------------------

func TestPauseSubscription_Success(t *testing.T) {
	svc := &mockSvc{
		pauseFn: func(_ string) error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/sub-xyz/pause", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

// ---- POST /subscriptions/{id}/resume -------------------------------------

func TestResumeSubscription_Success(t *testing.T) {
	svc := &mockSvc{
		resumeFn: func(_ string) error { return nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/subscriptions/sub-xyz/resume", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

// ---- GET /subscriptions/{id}/billing -------------------------------------

func TestListBillingRecords_Success(t *testing.T) {
	records := []*domain.BillingRecord{
		{ID: "br-1", SubscriptionID: "sub-001", Amount: 9.99, Currency: "USD", Status: "success", CreatedAt: time.Now()},
	}
	svc := &mockSvc{
		listBillingFn: func(subID string) ([]*domain.BillingRecord, error) {
			return records, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/subscriptions/sub-001/billing", nil)
	rr := httptest.NewRecorder()
	newHandler(svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp []map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp) != 1 {
		t.Errorf("expected 1 record, got %d", len(resp))
	}
}
