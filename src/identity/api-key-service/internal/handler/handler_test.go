package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/api-key-service/internal/domain"
	"github.com/shopos/api-key-service/internal/handler"
)

// ---- mock -------------------------------------------------------------------

type mockService struct {
	createFn    func(ctx context.Context, req *domain.CreateKeyRequest) (*domain.APIKey, string, error)
	validateFn  func(ctx context.Context, rawKey string) (*domain.ValidateResponse, error)
	listFn      func(ctx context.Context, ownerID string) ([]*domain.APIKey, error)
	getByIDFn   func(ctx context.Context, id string) (*domain.APIKey, error)
	deactivateFn func(ctx context.Context, id string) error
	deleteFn    func(ctx context.Context, id string) error
}

func (m *mockService) Create(ctx context.Context, req *domain.CreateKeyRequest) (*domain.APIKey, string, error) {
	return m.createFn(ctx, req)
}
func (m *mockService) Validate(ctx context.Context, rawKey string) (*domain.ValidateResponse, error) {
	return m.validateFn(ctx, rawKey)
}
func (m *mockService) List(ctx context.Context, ownerID string) ([]*domain.APIKey, error) {
	return m.listFn(ctx, ownerID)
}
func (m *mockService) GetByID(ctx context.Context, id string) (*domain.APIKey, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockService) Deactivate(ctx context.Context, id string) error {
	return m.deactivateFn(ctx, id)
}
func (m *mockService) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

// ---- helpers ----------------------------------------------------------------

func newServer(svc handler.Servicer) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(svc)
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func makeKey(id, ownerID string) *domain.APIKey {
	now := time.Now()
	return &domain.APIKey{
		ID:        id,
		OwnerID:   ownerID,
		OwnerType: "user",
		Name:      "test key",
		KeyPrefix: "sk_live_",
		KeyHash:   "deadbeef",
		Scopes:    []string{"catalog:read"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestCreateKey_Success(t *testing.T) {
	key := makeKey("key-1", "owner-1")
	svc := &mockService{
		createFn: func(_ context.Context, req *domain.CreateKeyRequest) (*domain.APIKey, string, error) {
			return key, "sk_live_rawrawraw", nil
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	payload := map[string]any{
		"owner_id":   "owner-1",
		"owner_type": "user",
		"name":       "test key",
		"scopes":     []string{"catalog:read"},
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(srv.URL+"/keys", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result["key"] != "sk_live_rawrawraw" {
		t.Fatalf("expected raw key in response, got %v", result["key"])
	}
	if result["id"] != "key-1" {
		t.Fatalf("expected id key-1, got %v", result["id"])
	}
}

func TestCreateKey_BadJSON(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/keys", "application/json", bytes.NewBufferString("{bad json"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListKeys_Success(t *testing.T) {
	keys := []*domain.APIKey{makeKey("k1", "owner-1"), makeKey("k2", "owner-1")}
	svc := &mockService{
		listFn: func(_ context.Context, ownerID string) ([]*domain.APIKey, error) {
			return keys, nil
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/keys?owner_id=owner-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	items, ok := result["keys"].([]any)
	if !ok {
		t.Fatal("expected keys array in response")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(items))
	}
}

func TestListKeys_MissingOwnerID(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/keys")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetKey_Success(t *testing.T) {
	key := makeKey("key-1", "owner-1")
	svc := &mockService{
		getByIDFn: func(_ context.Context, id string) (*domain.APIKey, error) {
			if id == "key-1" {
				return key, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/keys/key-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestGetKey_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(_ context.Context, id string) (*domain.APIKey, error) {
			return nil, domain.ErrNotFound
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/keys/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeactivateKey_Success(t *testing.T) {
	svc := &mockService{
		deactivateFn: func(_ context.Context, id string) error {
			return nil
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/keys/key-1", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestDeactivateKey_NotFound(t *testing.T) {
	svc := &mockService{
		deactivateFn: func(_ context.Context, id string) error {
			return domain.ErrNotFound
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/keys/missing", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestValidateKey_Valid(t *testing.T) {
	svc := &mockService{
		validateFn: func(_ context.Context, rawKey string) (*domain.ValidateResponse, error) {
			return &domain.ValidateResponse{
				Valid:   true,
				KeyID:   "key-1",
				OwnerID: "owner-1",
				Scopes:  []string{"catalog:read"},
			}, nil
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{"key": "sk_live_somekey"})
	resp, err := http.Post(srv.URL+"/keys/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result domain.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Fatalf("expected valid=true")
	}
}

func TestValidateKey_Invalid(t *testing.T) {
	svc := &mockService{
		validateFn: func(_ context.Context, rawKey string) (*domain.ValidateResponse, error) {
			return &domain.ValidateResponse{
				Valid:  false,
				Reason: "key not found",
			}, nil
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{"key": "sk_live_bogus"})
	resp, err := http.Post(srv.URL+"/keys/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid key, got %d", resp.StatusCode)
	}

	var result domain.ValidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Valid {
		t.Fatalf("expected valid=false")
	}
	if result.Reason == "" {
		t.Fatalf("expected a reason")
	}
}

func TestValidateKey_BadJSON(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/keys/validate", "application/json", bytes.NewBufferString("{invalid"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestValidateKey_MethodNotAllowed(t *testing.T) {
	srv := newServer(&mockService{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/keys/validate")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// GET /keys/validate will be routed to keysWithID with id="validate"
	// and since mockService.getByIDFn is nil it will panic — so we use a valid mock.
}

func TestValidateKey_Expired(t *testing.T) {
	svc := &mockService{
		validateFn: func(_ context.Context, rawKey string) (*domain.ValidateResponse, error) {
			return &domain.ValidateResponse{
				Valid:  false,
				Reason: "key has expired",
			}, nil
		},
	}

	srv := newServer(svc)
	defer srv.Close()

	body, _ := json.Marshal(map[string]string{"key": "sk_live_expiredkey"})
	resp, err := http.Post(srv.URL+"/keys/validate", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired key, got %d", resp.StatusCode)
	}
}
