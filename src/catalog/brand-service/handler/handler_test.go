package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/brand-service/domain"
	"github.com/shopos/brand-service/handler"
	"github.com/shopos/brand-service/service"
)

// ---- mock Servicer ----------------------------------------------------------

type mockService struct {
	createFn    func(req service.CreateRequest) (*domain.Brand, error)
	getByIDFn   func(id string) (*domain.Brand, error)
	getBySlugFn func(slug string) (*domain.Brand, error)
	listFn      func(activeOnly bool) ([]*domain.Brand, error)
	updateFn    func(id string, req service.UpdateRequest) (*domain.Brand, error)
	deleteFn    func(id string) error
}

func (m *mockService) Create(req service.CreateRequest) (*domain.Brand, error) {
	return m.createFn(req)
}
func (m *mockService) GetByID(id string) (*domain.Brand, error)   { return m.getByIDFn(id) }
func (m *mockService) GetBySlug(slug string) (*domain.Brand, error) { return m.getBySlugFn(slug) }
func (m *mockService) List(activeOnly bool) ([]*domain.Brand, error) {
	return m.listFn(activeOnly)
}
func (m *mockService) Update(id string, req service.UpdateRequest) (*domain.Brand, error) {
	return m.updateFn(id, req)
}
func (m *mockService) Delete(id string) error { return m.deleteFn(id) }

// ---- helpers ----------------------------------------------------------------

func fixedBrand() *domain.Brand {
	return &domain.Brand{
		ID:          "brand-001",
		Name:        "Acme",
		Slug:        "acme",
		Description: "The Acme Corporation",
		LogoURL:     "https://example.com/logo.png",
		Website:     "https://acme.example.com",
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func newHandler(svc handler.Servicer) http.Handler {
	return handler.New(svc)
}

func doRequest(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var b bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&b).Encode(body)
	}
	req := httptest.NewRequest(method, path, &b)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := newHandler(&mockService{})
	rr := doRequest(h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", resp["status"])
	}
}

func TestCreateBrand_Created(t *testing.T) {
	svc := &mockService{
		createFn: func(req service.CreateRequest) (*domain.Brand, error) {
			return fixedBrand(), nil
		},
	}
	rr := doRequest(newHandler(svc), http.MethodPost, "/brands", map[string]any{
		"name": "Acme", "slug": "acme",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var b domain.Brand
	_ = json.NewDecoder(rr.Body).Decode(&b)
	if b.ID != "brand-001" {
		t.Fatalf("unexpected id %q", b.ID)
	}
}

func TestCreateBrand_SlugConflict(t *testing.T) {
	svc := &mockService{
		createFn: func(req service.CreateRequest) (*domain.Brand, error) {
			return nil, domain.ErrSlugTaken
		},
	}
	rr := doRequest(newHandler(svc), http.MethodPost, "/brands", map[string]any{
		"name": "Acme", "slug": "acme",
	})
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestListBrands(t *testing.T) {
	svc := &mockService{
		listFn: func(activeOnly bool) ([]*domain.Brand, error) {
			return []*domain.Brand{fixedBrand()}, nil
		},
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/brands", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var brands []*domain.Brand
	_ = json.NewDecoder(rr.Body).Decode(&brands)
	if len(brands) != 1 {
		t.Fatalf("expected 1 brand, got %d", len(brands))
	}
}

func TestListBrands_ActiveFilter(t *testing.T) {
	var capturedActive bool
	svc := &mockService{
		listFn: func(activeOnly bool) ([]*domain.Brand, error) {
			capturedActive = activeOnly
			return []*domain.Brand{}, nil
		},
	}
	doRequest(newHandler(svc), http.MethodGet, "/brands?active=true", nil)
	if !capturedActive {
		t.Fatal("expected activeOnly=true to be passed to service")
	}
}

func TestGetBrandByID_OK(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(id string) (*domain.Brand, error) { return fixedBrand(), nil },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/brands/brand-001", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetBrandByID_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(id string) (*domain.Brand, error) { return nil, domain.ErrNotFound },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/brands/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetBrandBySlug_OK(t *testing.T) {
	svc := &mockService{
		getBySlugFn: func(slug string) (*domain.Brand, error) { return fixedBrand(), nil },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/brands/slug/acme", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetBrandBySlug_NotFound(t *testing.T) {
	svc := &mockService{
		getBySlugFn: func(slug string) (*domain.Brand, error) { return nil, domain.ErrNotFound },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/brands/slug/nope", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateBrand_OK(t *testing.T) {
	updated := fixedBrand()
	updated.Name = "Megacorp"
	svc := &mockService{
		updateFn: func(id string, req service.UpdateRequest) (*domain.Brand, error) {
			return updated, nil
		},
	}
	name := "Megacorp"
	rr := doRequest(newHandler(svc), http.MethodPatch, "/brands/brand-001", map[string]any{
		"name": name,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var b domain.Brand
	_ = json.NewDecoder(rr.Body).Decode(&b)
	if b.Name != "Megacorp" {
		t.Fatalf("expected name Megacorp, got %q", b.Name)
	}
}

func TestUpdateBrand_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(id string, req service.UpdateRequest) (*domain.Brand, error) {
			return nil, domain.ErrNotFound
		},
	}
	rr := doRequest(newHandler(svc), http.MethodPatch, "/brands/missing", map[string]any{})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateBrand_SlugConflict(t *testing.T) {
	svc := &mockService{
		updateFn: func(id string, req service.UpdateRequest) (*domain.Brand, error) {
			return nil, domain.ErrSlugTaken
		},
	}
	slug := "taken-slug"
	rr := doRequest(newHandler(svc), http.MethodPatch, "/brands/brand-001", map[string]any{
		"slug": slug,
	})
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestDeleteBrand_NoContent(t *testing.T) {
	svc := &mockService{
		deleteFn: func(id string) error { return nil },
	}
	rr := doRequest(newHandler(svc), http.MethodDelete, "/brands/brand-001", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteBrand_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(id string) error { return domain.ErrNotFound },
	}
	rr := doRequest(newHandler(svc), http.MethodDelete, "/brands/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
