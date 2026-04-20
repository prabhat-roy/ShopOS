package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/category-service/domain"
	"github.com/shopos/category-service/handler"
	"github.com/shopos/category-service/service"
)

// ---- mock Servicer ----------------------------------------------------------

type mockService struct {
	createFn    func(req service.CreateRequest) (*domain.Category, error)
	getByIDFn   func(id string) (*domain.Category, error)
	getBySlugFn func(slug string) (*domain.Category, error)
	listFn      func(parentID *string, activeOnly bool) ([]*domain.Category, error)
	updateFn    func(id string, req service.UpdateRequest) (*domain.Category, error)
	deleteFn    func(id string) error
}

func (m *mockService) Create(req service.CreateRequest) (*domain.Category, error) {
	return m.createFn(req)
}
func (m *mockService) GetByID(id string) (*domain.Category, error) { return m.getByIDFn(id) }
func (m *mockService) GetBySlug(slug string) (*domain.Category, error) {
	return m.getBySlugFn(slug)
}
func (m *mockService) List(parentID *string, activeOnly bool) ([]*domain.Category, error) {
	return m.listFn(parentID, activeOnly)
}
func (m *mockService) Update(id string, req service.UpdateRequest) (*domain.Category, error) {
	return m.updateFn(id, req)
}
func (m *mockService) Delete(id string) error { return m.deleteFn(id) }

// ---- helpers ----------------------------------------------------------------

func fixedCat() *domain.Category {
	return &domain.Category{
		ID:          "cat-001",
		Name:        "Electronics",
		Slug:        "electronics",
		Description: "All things electronic",
		ImageURL:    "https://example.com/img.png",
		SortOrder:   1,
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

func TestCreateCategory_Created(t *testing.T) {
	svc := &mockService{
		createFn: func(req service.CreateRequest) (*domain.Category, error) {
			return fixedCat(), nil
		},
	}
	rr := doRequest(newHandler(svc), http.MethodPost, "/categories", map[string]any{
		"name": "Electronics", "slug": "electronics",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var cat domain.Category
	_ = json.NewDecoder(rr.Body).Decode(&cat)
	if cat.ID != "cat-001" {
		t.Fatalf("unexpected id %q", cat.ID)
	}
}

func TestCreateCategory_SlugConflict(t *testing.T) {
	svc := &mockService{
		createFn: func(req service.CreateRequest) (*domain.Category, error) {
			return nil, domain.ErrSlugTaken
		},
	}
	rr := doRequest(newHandler(svc), http.MethodPost, "/categories", map[string]any{
		"name": "Electronics", "slug": "electronics",
	})
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestListCategories(t *testing.T) {
	svc := &mockService{
		listFn: func(parentID *string, activeOnly bool) ([]*domain.Category, error) {
			return []*domain.Category{fixedCat()}, nil
		},
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/categories", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var cats []*domain.Category
	_ = json.NewDecoder(rr.Body).Decode(&cats)
	if len(cats) != 1 {
		t.Fatalf("expected 1 category, got %d", len(cats))
	}
}

func TestListCategories_ActiveFilter(t *testing.T) {
	var capturedActive bool
	svc := &mockService{
		listFn: func(parentID *string, activeOnly bool) ([]*domain.Category, error) {
			capturedActive = activeOnly
			return []*domain.Category{}, nil
		},
	}
	doRequest(newHandler(svc), http.MethodGet, "/categories?active=true", nil)
	if !capturedActive {
		t.Fatal("expected activeOnly=true to be passed to service")
	}
}

func TestListCategories_ParentFilter(t *testing.T) {
	var capturedParent *string
	svc := &mockService{
		listFn: func(parentID *string, activeOnly bool) ([]*domain.Category, error) {
			capturedParent = parentID
			return []*domain.Category{}, nil
		},
	}
	doRequest(newHandler(svc), http.MethodGet, "/categories?parent_id=abc", nil)
	if capturedParent == nil || *capturedParent != "abc" {
		t.Fatal("expected parent_id=abc to be passed to service")
	}
}

func TestGetCategoryByID_OK(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(id string) (*domain.Category, error) { return fixedCat(), nil },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/categories/cat-001", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetCategoryByID_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(id string) (*domain.Category, error) { return nil, domain.ErrNotFound },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/categories/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetCategoryBySlug_OK(t *testing.T) {
	svc := &mockService{
		getBySlugFn: func(slug string) (*domain.Category, error) { return fixedCat(), nil },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/categories/slug/electronics", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetCategoryBySlug_NotFound(t *testing.T) {
	svc := &mockService{
		getBySlugFn: func(slug string) (*domain.Category, error) { return nil, domain.ErrNotFound },
	}
	rr := doRequest(newHandler(svc), http.MethodGet, "/categories/slug/nope", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateCategory_OK(t *testing.T) {
	updated := fixedCat()
	updated.Name = "Tech"
	svc := &mockService{
		updateFn: func(id string, req service.UpdateRequest) (*domain.Category, error) {
			return updated, nil
		},
	}
	name := "Tech"
	rr := doRequest(newHandler(svc), http.MethodPatch, "/categories/cat-001", map[string]any{
		"name": name,
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var cat domain.Category
	_ = json.NewDecoder(rr.Body).Decode(&cat)
	if cat.Name != "Tech" {
		t.Fatalf("expected name Tech, got %q", cat.Name)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(id string, req service.UpdateRequest) (*domain.Category, error) {
			return nil, domain.ErrNotFound
		},
	}
	rr := doRequest(newHandler(svc), http.MethodPatch, "/categories/missing", map[string]any{})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateCategory_SlugConflict(t *testing.T) {
	svc := &mockService{
		updateFn: func(id string, req service.UpdateRequest) (*domain.Category, error) {
			return nil, domain.ErrSlugTaken
		},
	}
	slug := "taken-slug"
	rr := doRequest(newHandler(svc), http.MethodPatch, "/categories/cat-001", map[string]any{
		"slug": slug,
	})
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestDeleteCategory_NoContent(t *testing.T) {
	svc := &mockService{
		deleteFn: func(id string) error { return nil },
	}
	rr := doRequest(newHandler(svc), http.MethodDelete, "/categories/cat-001", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(id string) error { return domain.ErrNotFound },
	}
	rr := doRequest(newHandler(svc), http.MethodDelete, "/categories/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
