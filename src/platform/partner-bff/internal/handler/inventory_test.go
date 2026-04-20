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

func TestGetStock(t *testing.T) {
	svc := &mockInventoryService{
		stock: &domain.StockLevel{ProductID: "p1", Available: 50, Reserved: 5},
	}
	h := handler.NewInventoryHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/inventory/products/p1", nil)
	req.SetPathValue("id", "p1")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetStock(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.StockLevel
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.ProductID != "p1" || body.Available != 50 {
		t.Errorf("unexpected stock: %+v", body)
	}
}

func TestGetBulkStock(t *testing.T) {
	svc := &mockInventoryService{
		bulkStock: []*domain.StockLevel{
			{ProductID: "p1", Available: 50},
			{ProductID: "p2", Available: 10},
		},
	}
	h := handler.NewInventoryHandler(svc)

	body, _ := json.Marshal(map[string][]string{"product_ids": {"p1", "p2"}})
	req := httptest.NewRequest(http.MethodPost, "/inventory/bulk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetBulkStock(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Items []*domain.StockLevel `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 stock levels, got %d", len(resp.Items))
	}
}

func TestGetBulkStockMissingBody(t *testing.T) {
	svc := &mockInventoryService{}
	h := handler.NewInventoryHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/inventory/bulk", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetBulkStock(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
