package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/web-bff/internal/domain"
	"github.com/shopos/web-bff/internal/handler"
)

func TestListProducts_Success(t *testing.T) {
	svc := &mockCatalogService{
		listFn: func(_ context.Context, _ *domain.ListProductsRequest) (*domain.ProductList, error) {
			return &domain.ProductList{Items: []*domain.Product{{ID: "p1", Name: "Laptop"}}, Total: 1}, nil
		},
	}
	h := handler.NewCatalogHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	h.ListProducts(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	var body domain.ProductList
	json.NewDecoder(rr.Body).Decode(&body)
	if len(body.Items) != 1 {
		t.Errorf("expected 1 product, got %d", len(body.Items))
	}
}

func TestGetProduct_EmptyID_Returns400(t *testing.T) {
	h := handler.NewCatalogHandler(&mockCatalogService{})
	req := httptest.NewRequest(http.MethodGet, "/products/", nil)
	rr := httptest.NewRecorder()
	h.GetProduct(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestSearch_MissingQuery_Returns400(t *testing.T) {
	h := handler.NewCatalogHandler(&mockCatalogService{})
	req := httptest.NewRequest(http.MethodGet, "/search", nil)
	rr := httptest.NewRecorder()
	h.Search(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestListCategories_Success(t *testing.T) {
	svc := &mockCatalogService{
		categoriesFn: func(_ context.Context) ([]*domain.Category, error) {
			return []*domain.Category{{ID: "c1", Name: "Electronics"}}, nil
		},
	}
	h := handler.NewCatalogHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/categories", nil)
	rr := httptest.NewRecorder()
	h.ListCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
