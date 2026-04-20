package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/compare-service/internal/domain"
	"github.com/shopos/compare-service/internal/handler"
)

// mockServicer is a test double for service.Servicer.
type mockServicer struct {
	addItemFn       func(ctx context.Context, customerID string, item domain.CompareItem) (*domain.CompareList, error)
	removeItemFn    func(ctx context.Context, customerID string, productID string) (*domain.CompareList, error)
	getCompareListFn func(ctx context.Context, customerID string) (*domain.CompareList, error)
	clearListFn     func(ctx context.Context, customerID string) error
}

func (m *mockServicer) AddItem(ctx context.Context, customerID string, item domain.CompareItem) (*domain.CompareList, error) {
	return m.addItemFn(ctx, customerID, item)
}
func (m *mockServicer) RemoveItem(ctx context.Context, customerID string, productID string) (*domain.CompareList, error) {
	return m.removeItemFn(ctx, customerID, productID)
}
func (m *mockServicer) GetCompareList(ctx context.Context, customerID string) (*domain.CompareList, error) {
	return m.getCompareListFn(ctx, customerID)
}
func (m *mockServicer) ClearList(ctx context.Context, customerID string) error {
	return m.clearListFn(ctx, customerID)
}

const testCustomerID = "cust-001"

func emptyList() *domain.CompareList {
	return &domain.CompareList{
		CustomerID: testCustomerID,
		Items:      []domain.CompareItem{},
		UpdatedAt:  time.Now().UTC(),
	}
}

func listWithItems(productIDs ...string) *domain.CompareList {
	items := make([]domain.CompareItem, 0, len(productIDs))
	for _, id := range productIDs {
		items = append(items, domain.CompareItem{
			ProductID:   id,
			ProductName: "Product " + id,
			Price:       99.99,
		})
	}
	return &domain.CompareList{
		CustomerID: testCustomerID,
		Items:      items,
		UpdatedAt:  time.Now().UTC(),
	}
}

// Test 1: GET /healthz returns 200 with status ok.
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

// Test 2: GET /compare/{customerId} returns 200 with list.
func TestGetList(t *testing.T) {
	mock := &mockServicer{
		getCompareListFn: func(_ context.Context, _ string) (*domain.CompareList, error) {
			return listWithItems("p1", "p2"), nil
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/compare/"+testCustomerID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list domain.CompareList
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
}

// Test 3: POST /compare/{customerId}/items returns 201 with updated list.
func TestAddItem_Created(t *testing.T) {
	mock := &mockServicer{
		addItemFn: func(_ context.Context, _ string, item domain.CompareItem) (*domain.CompareList, error) {
			return listWithItems(item.ProductID), nil
		},
	}
	h := handler.New(mock)

	body, _ := json.Marshal(domain.CompareItem{ProductID: "p1", ProductName: "Product 1", Price: 49.99})
	req := httptest.NewRequest(http.MethodPost, "/compare/"+testCustomerID+"/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

// Test 4: POST returns 422 when list is full.
func TestAddItem_ListFull(t *testing.T) {
	mock := &mockServicer{
		addItemFn: func(_ context.Context, _ string, _ domain.CompareItem) (*domain.CompareList, error) {
			return nil, domain.ErrListFull
		},
	}
	h := handler.New(mock)

	body, _ := json.Marshal(domain.CompareItem{ProductID: "p5", ProductName: "Product 5"})
	req := httptest.NewRequest(http.MethodPost, "/compare/"+testCustomerID+"/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", w.Code)
	}
}

// Test 5: DELETE /compare/{customerId}/items/{productId} returns 200 with updated list.
func TestRemoveItem_OK(t *testing.T) {
	mock := &mockServicer{
		removeItemFn: func(_ context.Context, _ string, _ string) (*domain.CompareList, error) {
			return emptyList(), nil
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodDelete, "/compare/"+testCustomerID+"/items/p1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// Test 6: DELETE /compare/{customerId}/items/{productId} returns 404 when not found.
func TestRemoveItem_NotFound(t *testing.T) {
	mock := &mockServicer{
		removeItemFn: func(_ context.Context, _ string, _ string) (*domain.CompareList, error) {
			return nil, domain.ErrItemNotFound
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodDelete, "/compare/"+testCustomerID+"/items/no-such", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// Test 7: DELETE /compare/{customerId} (clear) returns 204.
func TestClearList_NoContent(t *testing.T) {
	mock := &mockServicer{
		clearListFn: func(_ context.Context, _ string) error { return nil },
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodDelete, "/compare/"+testCustomerID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// Test 8: POST with invalid JSON body returns 400.
func TestAddItem_BadBody(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodPost, "/compare/"+testCustomerID+"/items", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
