package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/wishlist-service/internal/domain"
	"github.com/shopos/wishlist-service/internal/handler"
)

// mockServicer is a test double for service.Servicer.
type mockServicer struct {
	addToWishlistFn    func(ctx context.Context, customerID uuid.UUID, req *domain.AddItemRequest) (*domain.WishlistItem, error)
	getWishlistItemFn  func(ctx context.Context, customerID uuid.UUID, productID string) (*domain.WishlistItem, error)
	removeFromWishFn   func(ctx context.Context, customerID uuid.UUID, productID string) error
	getWishlistFn      func(ctx context.Context, customerID uuid.UUID, limit, offset int) (*domain.WishlistPage, error)
	clearWishlistFn    func(ctx context.Context, customerID uuid.UUID) error
	checkWishlistFn    func(ctx context.Context, customerID uuid.UUID, productID string) (bool, error)
}

func (m *mockServicer) AddToWishlist(ctx context.Context, customerID uuid.UUID, req *domain.AddItemRequest) (*domain.WishlistItem, error) {
	return m.addToWishlistFn(ctx, customerID, req)
}
func (m *mockServicer) GetWishlistItem(ctx context.Context, customerID uuid.UUID, productID string) (*domain.WishlistItem, error) {
	return m.getWishlistItemFn(ctx, customerID, productID)
}
func (m *mockServicer) RemoveFromWishlist(ctx context.Context, customerID uuid.UUID, productID string) error {
	return m.removeFromWishFn(ctx, customerID, productID)
}
func (m *mockServicer) GetWishlist(ctx context.Context, customerID uuid.UUID, limit, offset int) (*domain.WishlistPage, error) {
	return m.getWishlistFn(ctx, customerID, limit, offset)
}
func (m *mockServicer) ClearWishlist(ctx context.Context, customerID uuid.UUID) error {
	return m.clearWishlistFn(ctx, customerID)
}
func (m *mockServicer) CheckWishlist(ctx context.Context, customerID uuid.UUID, productID string) (bool, error) {
	return m.checkWishlistFn(ctx, customerID, productID)
}

var (
	testCustomerID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testItemID     = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testProductID  = "prod-abc-123"
)

func sampleItem() *domain.WishlistItem {
	return &domain.WishlistItem{
		ID:          testItemID,
		CustomerID:  testCustomerID,
		ProductID:   testProductID,
		ProductName: "Test Product",
		Price:       29.99,
		ImageURL:    "https://cdn.example.com/img.jpg",
		AddedAt:     time.Now().UTC(),
	}
}

// Test 1: GET /healthz returns 200 {"status":"ok"}
func TestHealthz(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", body["status"])
	}
}

// Test 2: POST /wishlist/{customerId}/items → 201 with item
func TestAddItem_Created(t *testing.T) {
	mock := &mockServicer{
		addToWishlistFn: func(_ context.Context, _ uuid.UUID, _ *domain.AddItemRequest) (*domain.WishlistItem, error) {
			return sampleItem(), nil
		},
	}
	h := handler.New(mock)

	reqBody, _ := json.Marshal(domain.AddItemRequest{
		ProductID:   testProductID,
		ProductName: "Test Product",
		Price:       29.99,
	})
	req := httptest.NewRequest(http.MethodPost, "/wishlist/"+testCustomerID.String()+"/items", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var item domain.WishlistItem
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.ProductID != testProductID {
		t.Fatalf("expected productId=%s, got %s", testProductID, item.ProductID)
	}
}

// Test 3: POST with duplicate → 409 Conflict
func TestAddItem_Conflict(t *testing.T) {
	mock := &mockServicer{
		addToWishlistFn: func(_ context.Context, _ uuid.UUID, _ *domain.AddItemRequest) (*domain.WishlistItem, error) {
			return nil, domain.ErrAlreadyExists
		},
	}
	h := handler.New(mock)

	reqBody, _ := json.Marshal(domain.AddItemRequest{ProductID: testProductID, ProductName: "Test"})
	req := httptest.NewRequest(http.MethodPost, "/wishlist/"+testCustomerID.String()+"/items", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

// Test 4: GET /wishlist/{customerId}/items → 200 with page
func TestListItems(t *testing.T) {
	mock := &mockServicer{
		getWishlistFn: func(_ context.Context, customerID uuid.UUID, limit, offset int) (*domain.WishlistPage, error) {
			return &domain.WishlistPage{
				CustomerID: customerID,
				Items:      []*domain.WishlistItem{sampleItem()},
				Total:      1,
				Limit:      limit,
				Offset:     offset,
			}, nil
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/wishlist/"+testCustomerID.String()+"/items?limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var page domain.WishlistPage
	if err := json.NewDecoder(w.Body).Decode(&page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Fatalf("expected total=1, got %d", page.Total)
	}
}

// Test 5: GET /wishlist/{customerId}/items/{productId} → 200
func TestGetItem_Found(t *testing.T) {
	mock := &mockServicer{
		getWishlistItemFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.WishlistItem, error) {
			return sampleItem(), nil
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/wishlist/"+testCustomerID.String()+"/items/"+testProductID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// Test 6: GET /wishlist/{customerId}/items/{productId} → 404 when not found
func TestGetItem_NotFound(t *testing.T) {
	mock := &mockServicer{
		getWishlistItemFn: func(_ context.Context, _ uuid.UUID, _ string) (*domain.WishlistItem, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/wishlist/"+testCustomerID.String()+"/items/no-such-product", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// Test 7: DELETE /wishlist/{customerId}/items/{productId} → 204
func TestRemoveItem_NoContent(t *testing.T) {
	mock := &mockServicer{
		removeFromWishFn: func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodDelete, "/wishlist/"+testCustomerID.String()+"/items/"+testProductID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// Test 8: DELETE /wishlist/{customerId} (clear) → 204
func TestClearWishlist_NoContent(t *testing.T) {
	mock := &mockServicer{
		clearWishlistFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodDelete, "/wishlist/"+testCustomerID.String(), nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// Test 9: GET /wishlist/{customerId}/check?productId= → true
func TestCheckWishlist_True(t *testing.T) {
	mock := &mockServicer{
		checkWishlistFn: func(_ context.Context, _ uuid.UUID, _ string) (bool, error) { return true, nil },
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/wishlist/"+testCustomerID.String()+"/check?productId="+testProductID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]bool
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body["inWishlist"] {
		t.Fatal("expected inWishlist=true")
	}
}

// Test 10: GET /wishlist/{customerId}/check without productId → 400
func TestCheckWishlist_MissingProductID(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodGet, "/wishlist/"+testCustomerID.String()+"/check", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
