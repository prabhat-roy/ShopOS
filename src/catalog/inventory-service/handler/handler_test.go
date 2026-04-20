package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/inventory-service/domain"
	"github.com/shopos/inventory-service/handler"
)

// -----------------------------------------------------------------------
// Mock Servicer
// -----------------------------------------------------------------------

type mockService struct {
	getStockFn      func(ctx context.Context, productID, warehouseID string) (*domain.StockLevel, error)
	listStockFn     func(ctx context.Context, productID string) ([]*domain.StockLevel, error)
	upsertStockFn   func(ctx context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error)
	reserveFn       func(ctx context.Context, orderID, productID string, qty int) (*domain.Reservation, error)
	releaseFn       func(ctx context.Context, reservationID string) error
	commitFn        func(ctx context.Context, reservationID string) error
	getReservationFn func(ctx context.Context, id string) (*domain.Reservation, error)
}

func (m *mockService) GetStock(ctx context.Context, productID, warehouseID string) (*domain.StockLevel, error) {
	return m.getStockFn(ctx, productID, warehouseID)
}
func (m *mockService) ListStock(ctx context.Context, productID string) ([]*domain.StockLevel, error) {
	return m.listStockFn(ctx, productID)
}
func (m *mockService) UpsertStock(ctx context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error) {
	return m.upsertStockFn(ctx, productID, sku, warehouseID, available, reorder)
}
func (m *mockService) Reserve(ctx context.Context, orderID, productID string, qty int) (*domain.Reservation, error) {
	return m.reserveFn(ctx, orderID, productID, qty)
}
func (m *mockService) Release(ctx context.Context, reservationID string) error {
	return m.releaseFn(ctx, reservationID)
}
func (m *mockService) Commit(ctx context.Context, reservationID string) error {
	return m.commitFn(ctx, reservationID)
}
func (m *mockService) GetReservation(ctx context.Context, id string) (*domain.Reservation, error) {
	return m.getReservationFn(ctx, id)
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

func newServer(svc handler.Servicer) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(svc)
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func sampleStock() *domain.StockLevel {
	return &domain.StockLevel{
		ID:          "sl-1",
		ProductID:   "prod-1",
		SKU:         "SKU-001",
		WarehouseID: "wh-1",
		Available:   100,
		Reserved:    5,
		Reorder:     10,
		UpdatedAt:   time.Now(),
	}
}

func sampleReservation() *domain.Reservation {
	return &domain.Reservation{
		ID:        "res-1",
		OrderID:   "ord-1",
		ProductID: "prod-1",
		Quantity:  3,
		Status:    domain.ReservedStatus,
		CreatedAt: time.Now(),
	}
}

// -----------------------------------------------------------------------
// GET /healthz
// -----------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %v", body)
	}
}

// -----------------------------------------------------------------------
// GET /inventory/{productID}
// -----------------------------------------------------------------------

func TestListStock_OK(t *testing.T) {
	svc := &mockService{
		listStockFn: func(_ context.Context, productID string) ([]*domain.StockLevel, error) {
			if productID != "prod-1" {
				t.Errorf("unexpected productID %s", productID)
			}
			return []*domain.StockLevel{sampleStock()}, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/inventory/prod-1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var levels []*domain.StockLevel
	json.NewDecoder(resp.Body).Decode(&levels)
	if len(levels) != 1 {
		t.Fatalf("expected 1 level, got %d", len(levels))
	}
}

func TestListStock_NotFound(t *testing.T) {
	svc := &mockService{
		listStockFn: func(_ context.Context, _ string) ([]*domain.StockLevel, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/inventory/prod-x")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// GET /inventory/{productID}/{warehouseID}
// -----------------------------------------------------------------------

func TestGetStock_OK(t *testing.T) {
	svc := &mockService{
		getStockFn: func(_ context.Context, productID, warehouseID string) (*domain.StockLevel, error) {
			return sampleStock(), nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/inventory/prod-1/wh-1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetStock_NotFound(t *testing.T) {
	svc := &mockService{
		getStockFn: func(_ context.Context, _, _ string) (*domain.StockLevel, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/inventory/prod-x/wh-x")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// PUT /inventory/{productID}/{warehouseID}
// -----------------------------------------------------------------------

func TestUpsertStock_OK(t *testing.T) {
	svc := &mockService{
		upsertStockFn: func(_ context.Context, productID, sku, warehouseID string, available, reorder int) (*domain.StockLevel, error) {
			sl := sampleStock()
			sl.Available = available
			sl.Reorder = reorder
			return sl, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	body := `{"sku":"SKU-001","available":50,"reorder_point":5}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/inventory/prod-1/wh-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// POST /inventory/reserve
// -----------------------------------------------------------------------

func TestReserve_OK(t *testing.T) {
	svc := &mockService{
		reserveFn: func(_ context.Context, _, _ string, _ int) (*domain.Reservation, error) {
			return sampleReservation(), nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	body := `{"order_id":"ord-1","product_id":"prod-1","quantity":3}`
	resp, err := http.Post(srv.URL+"/inventory/reserve", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestReserve_InsufficientStock_409(t *testing.T) {
	svc := &mockService{
		reserveFn: func(_ context.Context, _, _ string, _ int) (*domain.Reservation, error) {
			return nil, domain.ErrInsufficientStock
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	body := `{"order_id":"ord-1","product_id":"prod-1","quantity":9999}`
	resp, err := http.Post(srv.URL+"/inventory/reserve", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestReserve_BadRequest_MissingFields(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	body := `{"order_id":"ord-1"}`
	resp, err := http.Post(srv.URL+"/inventory/reserve", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// POST /inventory/release/{id}
// -----------------------------------------------------------------------

func TestRelease_OK(t *testing.T) {
	svc := &mockService{
		releaseFn: func(_ context.Context, id string) error {
			if id != "res-1" {
				return domain.ErrNotFound
			}
			return nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/inventory/release/res-1", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestRelease_NotFound(t *testing.T) {
	svc := &mockService{
		releaseFn: func(_ context.Context, _ string) error {
			return domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/inventory/release/bad-id", "application/json", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// POST /inventory/commit/{id}
// -----------------------------------------------------------------------

func TestCommit_OK(t *testing.T) {
	svc := &mockService{
		commitFn: func(_ context.Context, id string) error {
			if id != "res-1" {
				return domain.ErrNotFound
			}
			return nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/inventory/commit/res-1", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestCommit_NotFound(t *testing.T) {
	svc := &mockService{
		commitFn: func(_ context.Context, _ string) error {
			return domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/inventory/commit/bad-id", "application/json", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// Method-not-allowed guards
// -----------------------------------------------------------------------

func TestMethodNotAllowed_Reserve(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/inventory/reserve", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
