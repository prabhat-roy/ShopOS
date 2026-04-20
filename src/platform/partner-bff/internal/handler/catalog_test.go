package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/partner-bff/internal/domain"
	"github.com/shopos/partner-bff/internal/handler"
	"github.com/shopos/partner-bff/internal/service"
)

func TestListProducts(t *testing.T) {
	svc := &mockCatalogService{
		products: &domain.ProductList{
			Items: []*domain.Product{{ID: "p1", Name: "Widget", Price: 9.99}},
			Total: 1,
		},
	}
	h := handler.NewCatalogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/catalog/products", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.ListProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.ProductList
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "p1" {
		t.Errorf("unexpected products: %+v", body)
	}
}

func TestGetProduct(t *testing.T) {
	svc := &mockCatalogService{
		product: &domain.Product{ID: "p1", Name: "Widget", Price: 9.99},
	}
	h := handler.NewCatalogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/catalog/products/p1", nil)
	req.SetPathValue("id", "p1")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetProduct(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body domain.Product
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.ID != "p1" {
		t.Errorf("expected p1, got %q", body.ID)
	}
}

func TestListCategories(t *testing.T) {
	svc := &mockCatalogService{
		categories: []*domain.Category{{ID: "c1", Name: "Electronics"}},
	}
	h := handler.NewCatalogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/catalog/categories", nil)
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.ListCategories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Items []*domain.Category `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "c1" {
		t.Errorf("unexpected categories: %+v", body)
	}
}

func TestGetProductNotFound(t *testing.T) {
	svc := &mockCatalogService{err: service.ErrNotFound}
	h := handler.NewCatalogHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/catalog/products/missing", nil)
	req.SetPathValue("id", "missing")
	req.Header.Set("X-Partner-ID", "partnerA")
	w := httptest.NewRecorder()

	h.GetProduct(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
