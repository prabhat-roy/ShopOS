package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/warehouse-service/internal/domain"
	"github.com/shopos/warehouse-service/internal/handler"
)

// ---- mock -------------------------------------------------------------------

type mockService struct {
	warehouse   *domain.Warehouse
	warehouses  []*domain.Warehouse
	movement    *domain.StockMovement
	movements   []*domain.StockMovement
	stockLevel  int
	err         error
}

func (m *mockService) CreateWarehouse(w *domain.Warehouse) (*domain.Warehouse, error) {
	return m.warehouse, m.err
}
func (m *mockService) GetWarehouse(id string) (*domain.Warehouse, error) {
	return m.warehouse, m.err
}
func (m *mockService) ListWarehouses(_ bool) ([]*domain.Warehouse, error) {
	return m.warehouses, m.err
}
func (m *mockService) UpdateWarehouse(_ string, _ *domain.Warehouse) (*domain.Warehouse, error) {
	return m.warehouse, m.err
}
func (m *mockService) ReceiveStock(_ *domain.StockMovement) (*domain.StockMovement, error) {
	return m.movement, m.err
}
func (m *mockService) ShipStock(_ *domain.StockMovement) (*domain.StockMovement, error) {
	return m.movement, m.err
}
func (m *mockService) GetStockLevel(_, _ string) (int, error) {
	return m.stockLevel, m.err
}
func (m *mockService) ListMovements(_ string, _ int) ([]*domain.StockMovement, error) {
	return m.movements, m.err
}

// ---- helpers ----------------------------------------------------------------

func newServer(svc *mockService) http.Handler {
	return handler.New(svc)
}

func sampleWarehouse() *domain.Warehouse {
	return &domain.Warehouse{
		ID:        "wh-001",
		Name:      "Main DC",
		Location:  "US-EAST",
		Address:   "123 Dock St",
		Capacity:  1000,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func sampleMovement() *domain.StockMovement {
	return &domain.StockMovement{
		ID:           "mv-001",
		WarehouseID:  "wh-001",
		ProductID:    "prod-001",
		SKU:          "SKU-A",
		MovementType: domain.MovementInbound,
		Quantity:     10,
		ReferenceID:  "PO-001",
		Notes:        "",
		CreatedAt:    time.Now(),
	}
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	svc := &mockService{}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestCreateWarehouse_Success(t *testing.T) {
	wh := sampleWarehouse()
	svc := &mockService{warehouse: wh}
	srv := newServer(svc)

	payload, _ := json.Marshal(wh)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/warehouses", bytes.NewReader(payload))
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreateWarehouse_BadJSON(t *testing.T) {
	svc := &mockService{}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/warehouses", bytes.NewReader([]byte("{bad json")))
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListWarehouses(t *testing.T) {
	wh := sampleWarehouse()
	svc := &mockService{warehouses: []*domain.Warehouse{wh}}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/warehouses", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []*domain.Warehouse
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 warehouse, got %d", len(result))
	}
}

func TestGetWarehouse_NotFound(t *testing.T) {
	svc := &mockService{err: domain.ErrNotFound}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/warehouses/no-such-id", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateWarehouse_Success(t *testing.T) {
	wh := sampleWarehouse()
	svc := &mockService{warehouse: wh}
	srv := newServer(svc)

	payload, _ := json.Marshal(map[string]interface{}{"name": "Updated DC", "active": true})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/warehouses/wh-001", bytes.NewReader(payload))
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestReceiveStock_Success(t *testing.T) {
	mv := sampleMovement()
	svc := &mockService{movement: mv}
	srv := newServer(svc)

	payload, _ := json.Marshal(map[string]interface{}{
		"productId": "prod-001", "sku": "SKU-A", "quantity": 10, "referenceId": "PO-001",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/warehouses/wh-001/receive", bytes.NewReader(payload))
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestShipStock_InsufficientStock(t *testing.T) {
	svc := &mockService{err: domain.ErrInsufficientStock}
	srv := newServer(svc)

	payload, _ := json.Marshal(map[string]interface{}{
		"productId": "prod-001", "sku": "SKU-A", "quantity": 999,
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/warehouses/wh-001/ship", bytes.NewReader(payload))
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestGetStock(t *testing.T) {
	svc := &mockService{stockLevel: 42}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/warehouses/wh-001/stock?productId=prod-001", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result["stock"] != 42 {
		t.Fatalf("expected stock 42, got %d", result["stock"])
	}
}

func TestGetStock_MissingProductId(t *testing.T) {
	svc := &mockService{}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/warehouses/wh-001/stock", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestListMovements(t *testing.T) {
	mv := sampleMovement()
	svc := &mockService{movements: []*domain.StockMovement{mv}}
	srv := newServer(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/warehouses/wh-001/movements", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []*domain.StockMovement
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 movement, got %d", len(result))
	}
}

// Ensure the mock satisfies the service.Servicer interface at compile time.
var _ interface {
	CreateWarehouse(w *domain.Warehouse) (*domain.Warehouse, error)
	GetWarehouse(id string) (*domain.Warehouse, error)
	ListWarehouses(_ bool) ([]*domain.Warehouse, error)
	UpdateWarehouse(_ string, _ *domain.Warehouse) (*domain.Warehouse, error)
	ReceiveStock(_ *domain.StockMovement) (*domain.StockMovement, error)
	ShipStock(_ *domain.StockMovement) (*domain.StockMovement, error)
	GetStockLevel(_, _ string) (int, error)
	ListMovements(_ string, _ int) ([]*domain.StockMovement, error)
} = (*mockService)(nil)

// Compile-time check: errors package imported.
var _ = errors.New
