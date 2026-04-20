package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/shopos/product-catalog-service/internal/domain"
	"github.com/shopos/product-catalog-service/internal/handler"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockService struct {
	createFn      func(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error)
	getFn         func(ctx context.Context, id string) (*domain.Product, error)
	getBySKUFn    func(ctx context.Context, sku string) (*domain.Product, error)
	listFn        func(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error)
	updateFn      func(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error)
	deleteFn      func(ctx context.Context, id string) error
}

func (m *mockService) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	return m.createFn(ctx, req)
}
func (m *mockService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return m.getFn(ctx, id)
}
func (m *mockService) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	return m.getBySKUFn(ctx, sku)
}
func (m *mockService) ListProducts(ctx context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error) {
	return m.listFn(ctx, req)
}
func (m *mockService) UpdateProduct(ctx context.Context, id string, req *domain.UpdateProductRequest) (*domain.Product, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockService) DeleteProduct(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestHandler(svc handler.Servicer) http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return handler.New(svc, logger)
}

func doRequest(t *testing.T, h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /healthz
// ---------------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := doRequest(t, h, http.MethodGet, "/healthz", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	decodeBody(t, w, &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %q", resp["status"])
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /products
// ---------------------------------------------------------------------------

func TestCreateProduct_Success(t *testing.T) {
	created := &domain.Product{ID: "uuid-1", SKU: "SKU-001", Name: "Widget", Status: domain.StatusDraft}

	svc := &mockService{
		createFn: func(_ context.Context, _ *domain.CreateProductRequest) (*domain.Product, error) {
			return created, nil
		},
	}
	h := newTestHandler(svc)

	body := map[string]interface{}{
		"sku":   "SKU-001",
		"name":  "Widget",
		"price": 9.99,
	}
	w := doRequest(t, h, http.MethodPost, "/products", body)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp domain.Product
	decodeBody(t, w, &resp)
	if resp.ID != "uuid-1" {
		t.Errorf("expected id uuid-1, got %q", resp.ID)
	}
}

func TestCreateProduct_DuplicateSKU(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, _ *domain.CreateProductRequest) (*domain.Product, error) {
			return nil, domain.ErrDuplicateSKU
		},
	}
	h := newTestHandler(svc)

	body := map[string]interface{}{"sku": "DUP", "name": "Dupe"}
	w := doRequest(t, h, http.MethodPost, "/products", body)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestCreateProduct_InvalidBody(t *testing.T) {
	h := newTestHandler(&mockService{})
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader("not-json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /products
// ---------------------------------------------------------------------------

func TestListProducts_Success(t *testing.T) {
	list := &domain.ProductList{
		Items:  []*domain.Product{{ID: "1", SKU: "A"}, {ID: "2", SKU: "B"}},
		Total:  2,
		Limit:  20,
		Offset: 0,
	}

	svc := &mockService{
		listFn: func(_ context.Context, _ *domain.ListProductsRequest) (*domain.ProductList, error) {
			return list, nil
		},
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodGet, "/products", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp domain.ProductList
	decodeBody(t, w, &resp)
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
}

func TestListProducts_WithQueryParams(t *testing.T) {
	var capturedReq *domain.ListProductsRequest
	svc := &mockService{
		listFn: func(_ context.Context, req *domain.ListProductsRequest) (*domain.ProductList, error) {
			capturedReq = req
			return &domain.ProductList{Items: []*domain.Product{}, Total: 0}, nil
		},
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodGet, "/products?category_id=cat1&brand_id=br1&status=active&min_price=10&max_price=100&limit=5&offset=10", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedReq == nil {
		t.Fatal("expected request to be captured")
	}
	if capturedReq.CategoryID != "cat1" {
		t.Errorf("expected CategoryID cat1, got %q", capturedReq.CategoryID)
	}
	if capturedReq.BrandID != "br1" {
		t.Errorf("expected BrandID br1, got %q", capturedReq.BrandID)
	}
	if capturedReq.Limit != 5 {
		t.Errorf("expected Limit 5, got %d", capturedReq.Limit)
	}
	if capturedReq.Offset != 10 {
		t.Errorf("expected Offset 10, got %d", capturedReq.Offset)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /products/{id}
// ---------------------------------------------------------------------------

func TestGetProduct_Success(t *testing.T) {
	product := &domain.Product{ID: "abc", SKU: "SKU-X", Status: domain.StatusActive}

	svc := &mockService{
		getFn: func(_ context.Context, id string) (*domain.Product, error) {
			if id == "abc" {
				return product, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodGet, "/products/abc", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp domain.Product
	decodeBody(t, w, &resp)
	if resp.ID != "abc" {
		t.Errorf("expected id abc, got %q", resp.ID)
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	svc := &mockService{
		getFn: func(_ context.Context, _ string) (*domain.Product, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodGet, "/products/missing", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /products/sku/{sku}
// ---------------------------------------------------------------------------

func TestGetProductBySKU_Success(t *testing.T) {
	product := &domain.Product{ID: "xyz", SKU: "MY-SKU", Status: domain.StatusActive}

	svc := &mockService{
		getBySKUFn: func(_ context.Context, sku string) (*domain.Product, error) {
			if sku == "MY-SKU" {
				return product, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodGet, "/products/sku/MY-SKU", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp domain.Product
	decodeBody(t, w, &resp)
	if resp.SKU != "MY-SKU" {
		t.Errorf("expected SKU MY-SKU, got %q", resp.SKU)
	}
}

func TestGetProductBySKU_NotFound(t *testing.T) {
	svc := &mockService{
		getBySKUFn: func(_ context.Context, _ string) (*domain.Product, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodGet, "/products/sku/UNKNOWN", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: PATCH /products/{id}
// ---------------------------------------------------------------------------

func TestUpdateProduct_Success(t *testing.T) {
	newName := "Updated Widget"
	updated := &domain.Product{ID: "abc", Name: newName, Status: domain.StatusActive}

	svc := &mockService{
		updateFn: func(_ context.Context, id string, _ *domain.UpdateProductRequest) (*domain.Product, error) {
			if id == "abc" {
				return updated, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := newTestHandler(svc)

	body := map[string]interface{}{"name": newName}
	w := doRequest(t, h, http.MethodPatch, "/products/abc", body)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp domain.Product
	decodeBody(t, w, &resp)
	if resp.Name != newName {
		t.Errorf("expected Name %q, got %q", newName, resp.Name)
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, _ string, _ *domain.UpdateProductRequest) (*domain.Product, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newTestHandler(svc)

	body := map[string]interface{}{"name": "X"}
	w := doRequest(t, h, http.MethodPatch, "/products/missing", body)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdateProduct_InvalidBody(t *testing.T) {
	h := newTestHandler(&mockService{})
	req := httptest.NewRequest(http.MethodPatch, "/products/abc", strings.NewReader("{bad json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /products/{id}
// ---------------------------------------------------------------------------

func TestDeleteProduct_Success(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodDelete, "/products/abc", nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteProduct_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ string) error { return domain.ErrNotFound },
	}
	h := newTestHandler(svc)

	w := doRequest(t, h, http.MethodDelete, "/products/missing", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
