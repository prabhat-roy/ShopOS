package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/handler"
)

func TestGetCart_Success(t *testing.T) {
	svc := &mockCartService{
		getFn: func(_ context.Context, _ string) (*domain.Cart, error) {
			return &domain.Cart{UserID: "u1", Items: []*domain.CartItem{}, Total: 0}, nil
		},
	}
	h := handler.NewCartHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/cart", nil)
	req.Header.Set("X-User-ID", "u1")
	rr := httptest.NewRecorder()
	h.GetCart(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAddItem_MissingProductID_Returns400(t *testing.T) {
	h := handler.NewCartHandler(&mockCartService{})
	body, _ := json.Marshal(map[string]int{"quantity": 2})
	req := httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.AddItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAddItem_ZeroQuantity_Returns400(t *testing.T) {
	h := handler.NewCartHandler(&mockCartService{})
	body, _ := json.Marshal(map[string]any{"product_id": "p1", "quantity": 0})
	req := httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.AddItem(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestClearCart_Success(t *testing.T) {
	svc := &mockCartService{
		clearFn: func(_ context.Context, _ string) error { return nil },
	}
	h := handler.NewCartHandler(svc)
	req := httptest.NewRequest(http.MethodDelete, "/cart", nil)
	req.Header.Set("X-User-ID", "u1")
	rr := httptest.NewRecorder()
	h.ClearCart(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
