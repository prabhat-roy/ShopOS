package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/bundle-service/domain"
	"github.com/shopos/bundle-service/handler"
)

// -----------------------------------------------------------------------
// Mock Servicer
// -----------------------------------------------------------------------

type mockService struct {
	createFn  func(ctx context.Context, b *domain.Bundle) (*domain.Bundle, error)
	getByIDFn func(ctx context.Context, id string) (*domain.Bundle, error)
	listFn    func(ctx context.Context, activeOnly bool) ([]*domain.Bundle, error)
	updateFn  func(ctx context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error)
	deleteFn  func(ctx context.Context, id string) error
}

func (m *mockService) Create(ctx context.Context, b *domain.Bundle) (*domain.Bundle, error) {
	return m.createFn(ctx, b)
}
func (m *mockService) GetByID(ctx context.Context, id string) (*domain.Bundle, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockService) List(ctx context.Context, activeOnly bool) ([]*domain.Bundle, error) {
	return m.listFn(ctx, activeOnly)
}
func (m *mockService) Update(ctx context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error) {
	return m.updateFn(ctx, id, patch)
}
func (m *mockService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
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

func sampleBundle() *domain.Bundle {
	return &domain.Bundle{
		ID:          "bundle-1",
		Name:        "Starter Pack",
		Description: "Everything you need",
		Price:       49.99,
		Currency:    "USD",
		Items: []domain.BundleItem{
			{ProductID: "prod-1", SKU: "SKU-001", Quantity: 2, Price: 19.99},
		},
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
// POST /bundles
// -----------------------------------------------------------------------

func TestCreateBundle_OK(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, b *domain.Bundle) (*domain.Bundle, error) {
			b.ID = "bundle-1"
			return b, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	body := `{"name":"Starter Pack","price":49.99,"currency":"USD","items":[]}`
	resp, err := http.Post(srv.URL+"/bundles", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var result domain.Bundle
	json.NewDecoder(resp.Body).Decode(&result)
	if result.ID != "bundle-1" {
		t.Fatalf("expected id bundle-1, got %s", result.ID)
	}
}

func TestCreateBundle_MissingName(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	body := `{"price":49.99}`
	resp, err := http.Post(srv.URL+"/bundles", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateBundle_InvalidJSON(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/bundles", "application/json", bytes.NewBufferString("{bad"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// GET /bundles
// -----------------------------------------------------------------------

func TestListBundles_All(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, activeOnly bool) ([]*domain.Bundle, error) {
			if activeOnly {
				t.Error("expected activeOnly=false")
			}
			return []*domain.Bundle{sampleBundle()}, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/bundles")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var bundles []*domain.Bundle
	json.NewDecoder(resp.Body).Decode(&bundles)
	if len(bundles) != 1 {
		t.Fatalf("expected 1 bundle, got %d", len(bundles))
	}
}

func TestListBundles_ActiveOnly(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, activeOnly bool) ([]*domain.Bundle, error) {
			if !activeOnly {
				t.Error("expected activeOnly=true")
			}
			return []*domain.Bundle{sampleBundle()}, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/bundles?active=true")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestListBundles_EmptySlice(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, _ bool) ([]*domain.Bundle, error) {
			return nil, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/bundles")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var bundles []*domain.Bundle
	json.NewDecoder(resp.Body).Decode(&bundles)
	if len(bundles) != 0 {
		t.Fatalf("expected empty array, got %v", bundles)
	}
}

// -----------------------------------------------------------------------
// GET /bundles/{id}
// -----------------------------------------------------------------------

func TestGetBundle_OK(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(_ context.Context, id string) (*domain.Bundle, error) {
			if id != "bundle-1" {
				return nil, domain.ErrNotFound
			}
			return sampleBundle(), nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/bundles/bundle-1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetBundle_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(_ context.Context, _ string) (*domain.Bundle, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/bundles/does-not-exist")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// PATCH /bundles/{id}
// -----------------------------------------------------------------------

func TestUpdateBundle_OK(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, id string, patch *domain.Bundle) (*domain.Bundle, error) {
			b := sampleBundle()
			b.Name = patch.Name
			return b, nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	body := `{"name":"Updated Pack","active":true}`
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/bundles/bundle-1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result domain.Bundle
	json.NewDecoder(resp.Body).Decode(&result)
	if result.Name != "Updated Pack" {
		t.Fatalf("expected name 'Updated Pack', got %s", result.Name)
	}
}

func TestUpdateBundle_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, _ string, _ *domain.Bundle) (*domain.Bundle, error) {
			return nil, domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	body := `{"name":"X"}`
	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/bundles/bad-id", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateBundle_InvalidJSON(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/bundles/bundle-1", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// DELETE /bundles/{id}
// -----------------------------------------------------------------------

func TestDeleteBundle_OK(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, id string) error {
			if id != "bundle-1" {
				return domain.ErrNotFound
			}
			return nil
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/bundles/bundle-1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestDeleteBundle_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ string) error {
			return domain.ErrNotFound
		},
	}
	srv := newServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/bundles/bad-id", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// -----------------------------------------------------------------------
// Method-not-allowed
// -----------------------------------------------------------------------

func TestMethodNotAllowed_Collection(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/bundles", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestMethodNotAllowed_Item(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/bundles/bundle-1", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}
