package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/recently-viewed-service/internal/domain"
	"github.com/shopos/recently-viewed-service/internal/handler"
)

// mockServicer is a test double for service.Servicer.
type mockServicer struct {
	recordViewFn        func(ctx context.Context, customerID string, item domain.ViewedItem) error
	getRecentlyViewedFn func(ctx context.Context, customerID string, limit int) (*domain.RecentlyViewedList, error)
	clearHistoryFn      func(ctx context.Context, customerID string) error
	getCountFn          func(ctx context.Context, customerID string) (int, error)
}

func (m *mockServicer) RecordView(ctx context.Context, customerID string, item domain.ViewedItem) error {
	return m.recordViewFn(ctx, customerID, item)
}
func (m *mockServicer) GetRecentlyViewed(ctx context.Context, customerID string, limit int) (*domain.RecentlyViewedList, error) {
	return m.getRecentlyViewedFn(ctx, customerID, limit)
}
func (m *mockServicer) ClearHistory(ctx context.Context, customerID string) error {
	return m.clearHistoryFn(ctx, customerID)
}
func (m *mockServicer) GetCount(ctx context.Context, customerID string) (int, error) {
	return m.getCountFn(ctx, customerID)
}

const testCustomerID = "cust-rvs-001"

func sampleItem(productID string) domain.ViewedItem {
	return domain.ViewedItem{
		ProductID:   productID,
		ProductName: "Product " + productID,
		ImageURL:    "https://cdn.example.com/" + productID + ".jpg",
		Price:       19.99,
		ViewedAt:    time.Now().UTC(),
	}
}

// Test 1: GET /healthz returns 200 {"status":"ok"}.
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

// Test 2: POST /recently-viewed/{customerId} returns 201 when item is recorded.
func TestRecordView_Created(t *testing.T) {
	mock := &mockServicer{
		recordViewFn: func(_ context.Context, _ string, _ domain.ViewedItem) error { return nil },
	}
	h := handler.New(mock)

	body, _ := json.Marshal(sampleItem("p-001"))
	req := httptest.NewRequest(http.MethodPost, "/recently-viewed/"+testCustomerID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

// Test 3: POST with missing productId returns 400.
func TestRecordView_MissingProductID(t *testing.T) {
	h := handler.New(&mockServicer{})

	item := domain.ViewedItem{ProductName: "No ID Product"}
	body, _ := json.Marshal(item)
	req := httptest.NewRequest(http.MethodPost, "/recently-viewed/"+testCustomerID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// Test 4: POST with invalid JSON returns 400.
func TestRecordView_BadBody(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodPost, "/recently-viewed/"+testCustomerID, bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// Test 5: GET /recently-viewed/{customerId} returns 200 with list.
func TestGetRecentlyViewed_OK(t *testing.T) {
	mock := &mockServicer{
		getRecentlyViewedFn: func(_ context.Context, customerID string, limit int) (*domain.RecentlyViewedList, error) {
			return &domain.RecentlyViewedList{
				CustomerID: customerID,
				Items:      []domain.ViewedItem{sampleItem("p-001"), sampleItem("p-002")},
				Total:      2,
			}, nil
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/recently-viewed/"+testCustomerID+"?limit=10", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list domain.RecentlyViewedList
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list.Items))
	}
}

// Test 6: GET with no limit param uses default (still 200).
func TestGetRecentlyViewed_DefaultLimit(t *testing.T) {
	mock := &mockServicer{
		getRecentlyViewedFn: func(_ context.Context, customerID string, limit int) (*domain.RecentlyViewedList, error) {
			return &domain.RecentlyViewedList{
				CustomerID: customerID,
				Items:      []domain.ViewedItem{},
				Total:      0,
			}, nil
		},
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodGet, "/recently-viewed/"+testCustomerID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// Test 7: DELETE /recently-viewed/{customerId} returns 204.
func TestClearHistory_NoContent(t *testing.T) {
	mock := &mockServicer{
		clearHistoryFn: func(_ context.Context, _ string) error { return nil },
	}
	h := handler.New(mock)

	req := httptest.NewRequest(http.MethodDelete, "/recently-viewed/"+testCustomerID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

// Test 8: Unsupported HTTP method returns 405.
func TestUnsupportedMethod(t *testing.T) {
	h := handler.New(&mockServicer{})
	req := httptest.NewRequest(http.MethodPut, "/recently-viewed/"+testCustomerID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
