package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/permission-service/internal/domain"
	"github.com/shopos/permission-service/internal/handler"
)

// ---- mock -------------------------------------------------------------------

type mockService struct {
	roles     map[string]*domain.Role
	userRoles map[string][]*domain.UserRole // keyed by userID
}

func newMock() *mockService {
	return &mockService{
		roles:     make(map[string]*domain.Role),
		userRoles: make(map[string][]*domain.UserRole),
	}
}

func (m *mockService) CreateRole(_ context.Context, name, description string, permissions []string) (*domain.Role, error) {
	for _, r := range m.roles {
		if r.Name == name {
			return nil, domain.ErrAlreadyAssigned // reuse for uniqueness test
		}
	}
	r := &domain.Role{
		ID:          "role-" + name,
		Name:        name,
		Description: description,
		Permissions: permissions,
	}
	m.roles[r.ID] = r
	return r, nil
}

func (m *mockService) GetRole(_ context.Context, id string) (*domain.Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return r, nil
}

func (m *mockService) ListRoles(_ context.Context) ([]*domain.Role, error) {
	out := make([]*domain.Role, 0, len(m.roles))
	for _, r := range m.roles {
		out = append(out, r)
	}
	return out, nil
}

func (m *mockService) DeleteRole(_ context.Context, id string) error {
	if _, ok := m.roles[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.roles, id)
	return nil
}

func (m *mockService) AssignRole(_ context.Context, userID, roleID string) error {
	for _, ur := range m.userRoles[userID] {
		if ur.RoleID == roleID {
			return domain.ErrAlreadyAssigned
		}
	}
	if _, ok := m.roles[roleID]; !ok {
		return domain.ErrNotFound
	}
	m.userRoles[userID] = append(m.userRoles[userID], &domain.UserRole{
		UserID:   userID,
		RoleID:   roleID,
		RoleName: m.roles[roleID].Name,
	})
	return nil
}

func (m *mockService) RevokeRole(_ context.Context, userID, roleID string) error {
	urs := m.userRoles[userID]
	for i, ur := range urs {
		if ur.RoleID == roleID {
			m.userRoles[userID] = append(urs[:i], urs[i+1:]...)
			return nil
		}
	}
	return domain.ErrNotFound
}

func (m *mockService) GetUserRoles(_ context.Context, userID string) ([]*domain.UserRole, error) {
	return m.userRoles[userID], nil
}

func (m *mockService) Check(_ context.Context, req domain.CheckRequest) domain.CheckResponse {
	for _, ur := range m.userRoles[req.UserID] {
		if r, ok := m.roles[ur.RoleID]; ok {
			for _, p := range r.Permissions {
				if p == req.Permission {
					return domain.CheckResponse{Allowed: true, Reason: "permission granted"}
				}
			}
		}
	}
	return domain.CheckResponse{Allowed: false, Reason: "user does not have permission"}
}

// ---- helpers ----------------------------------------------------------------

func newHandler(svc handler.Servicer) http.Handler {
	return handler.New(svc)
}

func doRequest(h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodGet, "/healthz", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	decodeBody(t, w, &resp)
	if resp["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", resp["status"])
	}
}

func TestHealthz_WrongMethod(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodPost, "/healthz", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestCreateRole_Success(t *testing.T) {
	h := newHandler(newMock())
	body := map[string]any{
		"name":        "admin",
		"description": "Full access",
		"permissions": []string{"order:read", "order:write"},
	}
	w := doRequest(h, http.MethodPost, "/roles", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body=%s", w.Code, w.Body.String())
	}
	var role domain.Role
	decodeBody(t, w, &role)
	if role.Name != "admin" {
		t.Fatalf("expected name=admin, got %q", role.Name)
	}
	if len(role.Permissions) != 2 {
		t.Fatalf("expected 2 permissions, got %d", len(role.Permissions))
	}
}

func TestCreateRole_MissingName(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodPost, "/roles", map[string]any{"description": "no name"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListRoles(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "viewer", Permissions: []string{}}
	h := newHandler(m)
	w := doRequest(h, http.MethodGet, "/roles", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var roles []*domain.Role
	decodeBody(t, w, &roles)
	if len(roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(roles))
	}
}

func TestListRoles_Empty(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodGet, "/roles", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var roles []*domain.Role
	decodeBody(t, w, &roles)
	if roles == nil || len(roles) != 0 {
		t.Fatalf("expected empty array, got %v", roles)
	}
}

func TestGetRole_Found(t *testing.T) {
	m := newMock()
	m.roles["abc"] = &domain.Role{ID: "abc", Name: "editor", Permissions: []string{"cms:write"}}
	h := newHandler(m)
	w := doRequest(h, http.MethodGet, "/roles/abc", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var role domain.Role
	decodeBody(t, w, &role)
	if role.ID != "abc" {
		t.Fatalf("expected id=abc, got %q", role.ID)
	}
}

func TestGetRole_NotFound(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodGet, "/roles/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteRole_Success(t *testing.T) {
	m := newMock()
	m.roles["del1"] = &domain.Role{ID: "del1", Name: "temp", Permissions: []string{}}
	h := newHandler(m)
	w := doRequest(h, http.MethodDelete, "/roles/del1", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	if _, ok := m.roles["del1"]; ok {
		t.Fatal("expected role to be deleted")
	}
}

func TestDeleteRole_NotFound(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodDelete, "/roles/ghost", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAssignRole_Success(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "viewer", Permissions: []string{}}
	h := newHandler(m)
	w := doRequest(h, http.MethodPost, "/users/user-1/roles", map[string]string{"role_id": "r1"})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestAssignRole_Duplicate(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "viewer", Permissions: []string{}}
	m.userRoles["user-1"] = []*domain.UserRole{{UserID: "user-1", RoleID: "r1"}}
	h := newHandler(m)
	w := doRequest(h, http.MethodPost, "/users/user-1/roles", map[string]string{"role_id": "r1"})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestAssignRole_RoleNotFound(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodPost, "/users/user-1/roles", map[string]string{"role_id": "missing"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRevokeRole_Success(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "viewer", Permissions: []string{}}
	m.userRoles["user-1"] = []*domain.UserRole{{UserID: "user-1", RoleID: "r1"}}
	h := newHandler(m)
	w := doRequest(h, http.MethodDelete, "/users/user-1/roles/r1", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestRevokeRole_NotFound(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodDelete, "/users/user-1/roles/ghost", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetUserRoles_Success(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "viewer", Permissions: []string{}}
	m.userRoles["user-1"] = []*domain.UserRole{{UserID: "user-1", RoleID: "r1", RoleName: "viewer"}}
	h := newHandler(m)
	w := doRequest(h, http.MethodGet, "/users/user-1/roles", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var roles []*domain.UserRole
	decodeBody(t, w, &roles)
	if len(roles) != 1 {
		t.Fatalf("expected 1 user-role, got %d", len(roles))
	}
}

func TestGetUserRoles_Empty(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodGet, "/users/user-99/roles", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var roles []*domain.UserRole
	decodeBody(t, w, &roles)
	if roles == nil || len(roles) != 0 {
		t.Fatalf("expected empty array, got %v", roles)
	}
}

func TestCheck_Allowed(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "admin", Permissions: []string{"order:read", "order:write"}}
	m.userRoles["user-1"] = []*domain.UserRole{{UserID: "user-1", RoleID: "r1"}}
	h := newHandler(m)
	w := doRequest(h, http.MethodPost, "/check", domain.CheckRequest{UserID: "user-1", Permission: "order:read"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.CheckResponse
	decodeBody(t, w, &resp)
	if !resp.Allowed {
		t.Fatalf("expected allowed=true, reason=%s", resp.Reason)
	}
}

func TestCheck_Denied(t *testing.T) {
	m := newMock()
	m.roles["r1"] = &domain.Role{ID: "r1", Name: "viewer", Permissions: []string{"order:read"}}
	m.userRoles["user-1"] = []*domain.UserRole{{UserID: "user-1", RoleID: "r1"}}
	h := newHandler(m)
	w := doRequest(h, http.MethodPost, "/check", domain.CheckRequest{UserID: "user-1", Permission: "order:write"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.CheckResponse
	decodeBody(t, w, &resp)
	if resp.Allowed {
		t.Fatal("expected allowed=false")
	}
}

func TestCheck_NoRoles(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodPost, "/check", domain.CheckRequest{UserID: "nobody", Permission: "order:read"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp domain.CheckResponse
	decodeBody(t, w, &resp)
	if resp.Allowed {
		t.Fatal("expected allowed=false for user with no roles")
	}
}

func TestCheck_WrongMethod(t *testing.T) {
	h := newHandler(newMock())
	w := doRequest(h, http.MethodGet, "/check", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
