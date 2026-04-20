package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/tenant-service/internal/domain"
	"github.com/shopos/tenant-service/internal/handler"
)

// — mock Servicer —

type mockService struct {
	createFn         func(ctx context.Context, req domain.CreateTenantRequest) (*domain.Tenant, error)
	getFn            func(ctx context.Context, id string) (*domain.Tenant, error)
	getBySlugFn      func(ctx context.Context, slug string) (*domain.Tenant, error)
	listFn           func(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error)
	updateFn         func(ctx context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error)
	deleteFn         func(ctx context.Context, id string) error
	getSettingsFn    func(ctx context.Context, id string) (map[string]string, error)
	updateSettingsFn func(ctx context.Context, id string, settings map[string]string) (*domain.Tenant, error)
}

func (m *mockService) Create(ctx context.Context, req domain.CreateTenantRequest) (*domain.Tenant, error) {
	return m.createFn(ctx, req)
}
func (m *mockService) Get(ctx context.Context, id string) (*domain.Tenant, error) {
	return m.getFn(ctx, id)
}
func (m *mockService) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return m.getBySlugFn(ctx, slug)
}
func (m *mockService) List(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error) {
	return m.listFn(ctx, status, limit, offset)
}
func (m *mockService) Update(ctx context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error) {
	return m.updateFn(ctx, id, req)
}
func (m *mockService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}
func (m *mockService) GetSettings(ctx context.Context, id string) (map[string]string, error) {
	return m.getSettingsFn(ctx, id)
}
func (m *mockService) UpdateSettings(ctx context.Context, id string, settings map[string]string) (*domain.Tenant, error) {
	return m.updateSettingsFn(ctx, id, settings)
}

// — test helpers —

func newHandler(svc handler.Servicer) http.Handler {
	return handler.New(svc)
}

func doRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func decodeJSON(t *testing.T, rr *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.NewDecoder(rr.Body).Decode(dst); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
}

// — tests —

func TestHealthz(t *testing.T) {
	h := newHandler(&mockService{})
	rr := doRequest(t, h, http.MethodGet, "/healthz", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	decodeJSON(t, rr, &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestCreateTenant_Success(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, req domain.CreateTenantRequest) (*domain.Tenant, error) {
			return &domain.Tenant{
				ID:         "abc-123",
				Name:       req.Name,
				Slug:       req.Slug,
				OwnerEmail: req.OwnerEmail,
				Plan:       req.Plan,
				Status:     domain.StatusActive,
				Settings:   map[string]string{},
			}, nil
		},
	}
	h := newHandler(svc)

	body := domain.CreateTenantRequest{
		Name: "Acme Corp", Slug: "acme-corp",
		OwnerEmail: "admin@acme.com", Plan: domain.PlanPro,
	}
	rr := doRequest(t, h, http.MethodPost, "/tenants", body)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body)
	}

	var resp domain.Tenant
	decodeJSON(t, rr, &resp)
	if resp.ID != "abc-123" {
		t.Errorf("expected id=abc-123, got %q", resp.ID)
	}
	if resp.Slug != "acme-corp" {
		t.Errorf("expected slug=acme-corp, got %q", resp.Slug)
	}
}

func TestCreateTenant_SlugConflict(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, _ domain.CreateTenantRequest) (*domain.Tenant, error) {
			return nil, domain.ErrSlugTaken
		},
	}
	h := newHandler(svc)

	body := domain.CreateTenantRequest{
		Name: "Dupe", Slug: "existing", OwnerEmail: "x@x.com", Plan: domain.PlanStarter,
	}
	rr := doRequest(t, h, http.MethodPost, "/tenants", body)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestGetTenant_Success(t *testing.T) {
	want := &domain.Tenant{ID: "t-1", Name: "Test", Slug: "test", Status: domain.StatusActive}
	svc := &mockService{
		getFn: func(_ context.Context, id string) (*domain.Tenant, error) {
			if id == "t-1" {
				return want, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)

	rr := doRequest(t, h, http.MethodGet, "/tenants/t-1", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp domain.Tenant
	decodeJSON(t, rr, &resp)
	if resp.ID != "t-1" {
		t.Errorf("expected id=t-1, got %q", resp.ID)
	}
}

func TestGetTenant_NotFound(t *testing.T) {
	svc := &mockService{
		getFn: func(_ context.Context, _ string) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)

	rr := doRequest(t, h, http.MethodGet, "/tenants/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetTenantBySlug_Success(t *testing.T) {
	want := &domain.Tenant{ID: "t-2", Slug: "my-org", Status: domain.StatusActive}
	svc := &mockService{
		getBySlugFn: func(_ context.Context, slug string) (*domain.Tenant, error) {
			if slug == "my-org" {
				return want, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)

	rr := doRequest(t, h, http.MethodGet, "/tenants/slug/my-org", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
	var resp domain.Tenant
	decodeJSON(t, rr, &resp)
	if resp.Slug != "my-org" {
		t.Errorf("expected slug=my-org, got %q", resp.Slug)
	}
}

func TestGetTenantBySlug_NotFound(t *testing.T) {
	svc := &mockService{
		getBySlugFn: func(_ context.Context, _ string) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodGet, "/tenants/slug/nope", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestListTenants(t *testing.T) {
	svc := &mockService{
		listFn: func(_ context.Context, status string, limit, offset int) ([]*domain.Tenant, error) {
			return []*domain.Tenant{
				{ID: "t-1", Status: domain.StatusActive},
				{ID: "t-2", Status: domain.StatusActive},
			}, nil
		},
	}
	h := newHandler(svc)

	rr := doRequest(t, h, http.MethodGet, "/tenants?limit=10&offset=0", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []*domain.Tenant
	decodeJSON(t, rr, &resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(resp))
	}
}

func TestUpdateTenant_Success(t *testing.T) {
	name := "Updated Name"
	svc := &mockService{
		updateFn: func(_ context.Context, id string, req domain.UpdateTenantRequest) (*domain.Tenant, error) {
			return &domain.Tenant{ID: id, Name: *req.Name, Status: domain.StatusActive}, nil
		},
	}
	h := newHandler(svc)

	body := domain.UpdateTenantRequest{Name: &name}
	rr := doRequest(t, h, http.MethodPatch, "/tenants/t-1", body)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
	var resp domain.Tenant
	decodeJSON(t, rr, &resp)
	if resp.Name != "Updated Name" {
		t.Errorf("expected name='Updated Name', got %q", resp.Name)
	}
}

func TestUpdateTenant_NotFound(t *testing.T) {
	svc := &mockService{
		updateFn: func(_ context.Context, _ string, _ domain.UpdateTenantRequest) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)
	name := "x"
	body := domain.UpdateTenantRequest{Name: &name}
	rr := doRequest(t, h, http.MethodPatch, "/tenants/missing", body)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDeleteTenant_Success(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ string) error { return nil },
	}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodDelete, "/tenants/t-1", nil)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteTenant_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_ context.Context, _ string) error { return domain.ErrNotFound },
	}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodDelete, "/tenants/missing", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetSettings_Success(t *testing.T) {
	settings := map[string]string{"theme": "dark", "locale": "en-US"}
	svc := &mockService{
		getSettingsFn: func(_ context.Context, id string) (map[string]string, error) {
			if id == "t-1" {
				return settings, nil
			}
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)

	rr := doRequest(t, h, http.MethodGet, "/tenants/t-1/settings", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
	var resp map[string]string
	decodeJSON(t, rr, &resp)
	if resp["theme"] != "dark" {
		t.Errorf("expected theme=dark, got %q", resp["theme"])
	}
}

func TestGetSettings_NotFound(t *testing.T) {
	svc := &mockService{
		getSettingsFn: func(_ context.Context, _ string) (map[string]string, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)
	rr := doRequest(t, h, http.MethodGet, "/tenants/missing/settings", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUpdateSettings_Success(t *testing.T) {
	svc := &mockService{
		updateSettingsFn: func(_ context.Context, id string, settings map[string]string) (*domain.Tenant, error) {
			return &domain.Tenant{ID: id, Settings: settings, Status: domain.StatusActive}, nil
		},
	}
	h := newHandler(svc)

	newSettings := map[string]string{"feature_x": "enabled"}
	rr := doRequest(t, h, http.MethodPut, "/tenants/t-1/settings", newSettings)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body)
	}
	var resp domain.Tenant
	decodeJSON(t, rr, &resp)
	if resp.Settings["feature_x"] != "enabled" {
		t.Errorf("expected feature_x=enabled, got %q", resp.Settings["feature_x"])
	}
}

func TestUpdateSettings_NotFound(t *testing.T) {
	svc := &mockService{
		updateSettingsFn: func(_ context.Context, _ string, _ map[string]string) (*domain.Tenant, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)
	body := map[string]string{"k": "v"}
	rr := doRequest(t, h, http.MethodPut, "/tenants/missing/settings", body)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestCreateTenant_InvalidBody(t *testing.T) {
	h := newHandler(&mockService{})
	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
