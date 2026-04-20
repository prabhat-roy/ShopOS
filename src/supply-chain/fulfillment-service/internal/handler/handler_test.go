package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/fulfillment-service/internal/domain"
	"github.com/shopos/fulfillment-service/internal/handler"
)

// ---- mock -------------------------------------------------------------------

type mockService struct {
	fulfillment  *domain.FulfillmentOrder
	fulfillments []*domain.FulfillmentOrder
	err          error
}

func (m *mockService) CreateFulfillment(_ *domain.FulfillmentOrder) (*domain.FulfillmentOrder, error) {
	return m.fulfillment, m.err
}
func (m *mockService) GetFulfillment(_ string) (*domain.FulfillmentOrder, error) {
	return m.fulfillment, m.err
}
func (m *mockService) ListFulfillments(_ string) ([]*domain.FulfillmentOrder, error) {
	return m.fulfillments, m.err
}
func (m *mockService) GetByOrderID(_ string) (*domain.FulfillmentOrder, error) {
	return m.fulfillment, m.err
}
func (m *mockService) StartPicking(_ string) error  { return m.err }
func (m *mockService) StartPacking(_ string) error  { return m.err }
func (m *mockService) MarkReadyToShip(_ string) error { return m.err }
func (m *mockService) Ship(_, _, _ string) error    { return m.err }
func (m *mockService) Deliver(_ string) error       { return m.err }
func (m *mockService) Cancel(_ string) error        { return m.err }

// ---- helpers ----------------------------------------------------------------

func newServer(svc *mockService) http.Handler {
	return handler.New(svc)
}

func sampleOrder() *domain.FulfillmentOrder {
	return &domain.FulfillmentOrder{
		ID:          "f-001",
		OrderID:     "ord-001",
		WarehouseID: "wh-001",
		Status:      domain.StatusPending,
		Items: []domain.FulfillmentItem{
			{ID: "fi-001", FulfillmentID: "f-001", ProductID: "p-001", SKU: "SKU-A", Quantity: 2},
		},
		ShippingAddress: "123 Main St",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	svc := &mockService{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestCreateFulfillment_Success(t *testing.T) {
	order := sampleOrder()
	svc := &mockService{fulfillment: order}
	payload, _ := json.Marshal(order)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments", bytes.NewReader(payload))
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateFulfillment_BadJSON(t *testing.T) {
	svc := &mockService{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments", bytes.NewReader([]byte("{bad")))
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListFulfillments(t *testing.T) {
	svc := &mockService{fulfillments: []*domain.FulfillmentOrder{sampleOrder()}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fulfillments", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []*domain.FulfillmentOrder
	_ = json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
}

func TestGetFulfillment_NotFound(t *testing.T) {
	svc := &mockService{err: domain.ErrNotFound}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fulfillments/no-such-id", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetByOrderID(t *testing.T) {
	svc := &mockService{fulfillment: sampleOrder()}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fulfillments/order/ord-001", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPickFulfillment(t *testing.T) {
	svc := &mockService{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments/f-001/pick", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestShipFulfillment_Success(t *testing.T) {
	svc := &mockService{}
	payload, _ := json.Marshal(map[string]string{"trackingNumber": "1Z999", "carrier": "UPS"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments/f-001/ship", bytes.NewReader(payload))
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestShipFulfillment_BadJSON(t *testing.T) {
	svc := &mockService{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments/f-001/ship", bytes.NewReader([]byte("{bad")))
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeliverFulfillment(t *testing.T) {
	svc := &mockService{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments/f-001/deliver", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestCancelFulfillment(t *testing.T) {
	svc := &mockService{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/fulfillments/f-001", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestInvalidTransition_Conflict(t *testing.T) {
	svc := &mockService{err: domain.ErrInvalidTransition}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/fulfillments/f-001/pack", nil)
	newServer(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// Compile-time interface check.
var _ interface {
	CreateFulfillment(_ *domain.FulfillmentOrder) (*domain.FulfillmentOrder, error)
	GetFulfillment(_ string) (*domain.FulfillmentOrder, error)
	ListFulfillments(_ string) ([]*domain.FulfillmentOrder, error)
	GetByOrderID(_ string) (*domain.FulfillmentOrder, error)
	StartPicking(_ string) error
	StartPacking(_ string) error
	MarkReadyToShip(_ string) error
	Ship(_, _, _ string) error
	Deliver(_ string) error
	Cancel(_ string) error
} = (*mockService)(nil)
