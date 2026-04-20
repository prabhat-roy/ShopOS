package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/returns-logistics-service/internal/domain"
	"github.com/shopos/returns-logistics-service/internal/handler"
)

// ─── mock servicer ────────────────────────────────────────────────────────────

type mockService struct {
	ra  *domain.ReturnAuth
	ras []*domain.ReturnAuth
	err error
}

func (m *mockService) CreateReturnAuth(_, _, _ string, _ []domain.ReturnItem) (*domain.ReturnAuth, error) {
	return m.ra, m.err
}
func (m *mockService) GetReturnAuth(_ string) (*domain.ReturnAuth, error) {
	return m.ra, m.err
}
func (m *mockService) ListReturnAuths(_ string) ([]*domain.ReturnAuth, error) {
	return m.ras, m.err
}
func (m *mockService) ApproveReturn(_ string) (*domain.ReturnAuth, error)          { return m.ra, m.err }
func (m *mockService) RejectReturn(_, _ string) (*domain.ReturnAuth, error)        { return m.ra, m.err }
func (m *mockService) IssueLabel(_ string) (*domain.ReturnAuth, error)             { return m.ra, m.err }
func (m *mockService) MarkInTransit(_ string) (*domain.ReturnAuth, error)          { return m.ra, m.err }
func (m *mockService) MarkReceived(_ string) (*domain.ReturnAuth, error)           { return m.ra, m.err }
func (m *mockService) StartInspection(_ string) (*domain.ReturnAuth, error)        { return m.ra, m.err }
func (m *mockService) CompleteReturn(_, _ string) (*domain.ReturnAuth, error)      { return m.ra, m.err }
func (m *mockService) Cancel(_ string) (*domain.ReturnAuth, error)                 { return m.ra, m.err }

// ─── helpers ─────────────────────────────────────────────────────────────────

func sampleRA() *domain.ReturnAuth {
	return &domain.ReturnAuth{
		ID:         "ra-001",
		OrderID:    "ord-100",
		CustomerID: "cust-42",
		Items: []domain.ReturnItem{
			{ProductID: "prod-1", SKU: "SKU-001", Quantity: 1, Condition: "USED"},
		},
		Reason:    "defective product",
		Status:    domain.StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

func newHandler(svc *mockService) http.Handler {
	return handler.NewWithServicer(svc)
}

func doRequest(t *testing.T, h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
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
	h := newHandler(&mockService{})
	rr := doRequest(t, h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected ok, got %q", resp["status"])
	}
}

func TestCreateReturn_Success(t *testing.T) {
	svc := &mockService{ra: sampleRA()}
	h := newHandler(svc)
	body := map[string]interface{}{
		"orderId":    "ord-100",
		"customerId": "cust-42",
		"reason":     "defective product",
		"items": []map[string]interface{}{
			{"productId": "prod-1", "sku": "SKU-001", "quantity": 1, "condition": "USED"},
		},
	}
	rr := doRequest(t, h, http.MethodPost, "/returns", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var ra domain.ReturnAuth
	if err := json.NewDecoder(rr.Body).Decode(&ra); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ra.ID == "" {
		t.Error("expected non-empty ID")
	}
	if ra.Status != domain.StatusPending {
		t.Errorf("expected PENDING, got %s", ra.Status)
	}
}

func TestCreateReturn_MissingFields(t *testing.T) {
	h := newHandler(&mockService{})
	body := map[string]interface{}{"reason": "test"}
	rr := doRequest(t, h, http.MethodPost, "/returns", body)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetReturn_Found(t *testing.T) {
	svc := &mockService{ra: sampleRA()}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodGet, "/returns/ra-001", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var ra domain.ReturnAuth
	if err := json.NewDecoder(rr.Body).Decode(&ra); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ra.ID != "ra-001" {
		t.Errorf("expected ra-001, got %s", ra.ID)
	}
}

func TestGetReturn_NotFound(t *testing.T) {
	svc := &mockService{err: domain.ErrNotFound}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodGet, "/returns/does-not-exist", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListReturns(t *testing.T) {
	svc := &mockService{ras: []*domain.ReturnAuth{sampleRA(), sampleRA()}}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodGet, "/returns", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if total, ok := resp["total"].(float64); !ok || int(total) != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestApproveReturn(t *testing.T) {
	approved := sampleRA()
	approved.Status = domain.StatusApproved
	svc := &mockService{ra: approved}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodPost, "/returns/ra-001/approve", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestRejectReturn(t *testing.T) {
	rejected := sampleRA()
	rejected.Status = domain.StatusRejected
	svc := &mockService{ra: rejected}
	h := newHandler(svc)
	body := map[string]string{"reason": "item is not returnable"}
	rr := doRequest(t, h, http.MethodPost, "/returns/ra-001/reject", body)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestIssueLabel(t *testing.T) {
	labeled := sampleRA()
	labeled.Status = domain.StatusLabelIssued
	labeled.TrackingNumber = "RET-ABCD1234"
	labeled.ReturnLabel = "https://returns.shopos.internal/labels/RET-ABCD1234.pdf"
	svc := &mockService{ra: labeled}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodPost, "/returns/ra-001/label", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["trackingNumber"] == "" {
		t.Error("expected non-empty trackingNumber")
	}
}

func TestCancelReturn(t *testing.T) {
	cancelled := sampleRA()
	cancelled.Status = domain.StatusCancelled
	svc := &mockService{ra: cancelled}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodDelete, "/returns/ra-001", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTransition_Conflict(t *testing.T) {
	svc := &mockService{err: fmt.Errorf("%w: COMPLETED → APPROVED", domain.ErrInvalidTransition)}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodPost, "/returns/ra-001/approve", nil)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}
