package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/handler"
)

func TestListOrders(t *testing.T) {
	svc := &mockOrderService{
		orders: []*domain.Order{
			{ID: "o1", PartnerID: "partnerA", Status: "pending", Total: 99.99},
		},
	}
	h := handler.NewOrderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.ListOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Order `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "o1" {
		t.Errorf("unexpected orders: %+v", body)
	}
}

func TestGetOrder(t *testing.T) {
	svc := &mockOrderService{
		order: &domain.Order{ID: "o1", PartnerID: "partnerA", Status: "pending"},
	}
	h := handler.NewOrderHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/orders/o1", nil)
	req.SetPathValue("id", "o1")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetOrder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.Order
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.ID != "o1" {
		t.Errorf("expected o1, got %q", body.ID)
	}
}

func TestPlaceOrder(t *testing.T) {
	svc := &mockOrderService{
		order: &domain.Order{ID: "o2", PartnerID: "partnerA", Status: "pending"},
	}
	h := handler.NewOrderHandler(svc)

	payload := domain.PlaceOrderRequest{
		Items:     []domain.OrderItem{{ProductID: "p1", Quantity: 2, Price: 10.0}},
		AddressID: "addr1",
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.PlaceOrder(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestPlaceOrderMissingAddress(t *testing.T) {
	svc := &mockOrderService{}
	h := handler.NewOrderHandler(svc)

	payload := domain.PlaceOrderRequest{
		Items: []domain.OrderItem{{ProductID: "p1", Quantity: 1}},
	}
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.PlaceOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPlaceOrderBadBody(t *testing.T) {
	svc := &mockOrderService{}
	h := handler.NewOrderHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.PlaceOrder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
