package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopos/gdpr-service/internal/domain"
	"github.com/shopos/gdpr-service/internal/handler"
)

// ---------- mock service ----------

type mockService struct {
	submitRequestFn  func(ctx context.Context, userID string, reqType domain.RequestType, reason string) (*domain.DataRequest, error)
	getRequestFn     func(ctx context.Context, id string) (*domain.DataRequest, error)
	listRequestsFn   func(ctx context.Context, userID string) ([]*domain.DataRequest, error)
	processRequestFn func(ctx context.Context, id string, notes string) error
	completeRequestFn func(ctx context.Context, id string, notes string) error
	rejectRequestFn  func(ctx context.Context, id string, notes string) error
	updateConsentFn  func(ctx context.Context, userID string, consentType domain.ConsentType, granted bool, ip string) error
	getConsentsFn    func(ctx context.Context, userID string) ([]*domain.Consent, error)
	checkConsentFn   func(ctx context.Context, userID string, consentType domain.ConsentType) (bool, error)
}

func (m *mockService) SubmitRequest(ctx context.Context, userID string, reqType domain.RequestType, reason string) (*domain.DataRequest, error) {
	return m.submitRequestFn(ctx, userID, reqType, reason)
}
func (m *mockService) GetRequest(ctx context.Context, id string) (*domain.DataRequest, error) {
	return m.getRequestFn(ctx, id)
}
func (m *mockService) ListRequests(ctx context.Context, userID string) ([]*domain.DataRequest, error) {
	return m.listRequestsFn(ctx, userID)
}
func (m *mockService) ProcessRequest(ctx context.Context, id string, notes string) error {
	return m.processRequestFn(ctx, id, notes)
}
func (m *mockService) CompleteRequest(ctx context.Context, id string, notes string) error {
	return m.completeRequestFn(ctx, id, notes)
}
func (m *mockService) RejectRequest(ctx context.Context, id string, notes string) error {
	return m.rejectRequestFn(ctx, id, notes)
}
func (m *mockService) UpdateConsent(ctx context.Context, userID string, consentType domain.ConsentType, granted bool, ip string) error {
	return m.updateConsentFn(ctx, userID, consentType, granted, ip)
}
func (m *mockService) GetConsents(ctx context.Context, userID string) ([]*domain.Consent, error) {
	return m.getConsentsFn(ctx, userID)
}
func (m *mockService) CheckConsent(ctx context.Context, userID string, consentType domain.ConsentType) (bool, error) {
	return m.checkConsentFn(ctx, userID, consentType)
}

// ---------- helpers ----------

func sampleRequest() *domain.DataRequest {
	now := time.Now().UTC()
	return &domain.DataRequest{
		ID:        "req-1",
		UserID:    "user-1",
		Type:      domain.RequestExport,
		Status:    domain.StatusPending,
		Reason:    "I want my data",
		Notes:     "",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newHandler(svc handler.Servicer) http.Handler {
	return handler.New(svc)
}

func do(t *testing.T, h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// ---------- tests ----------

func TestHealthz(t *testing.T) {
	h := newHandler(&mockService{})
	rec := do(t, h, http.MethodGet, "/healthz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("want status=ok, got %q", resp["status"])
	}
}

func TestSubmitRequest_Created(t *testing.T) {
	svc := &mockService{
		submitRequestFn: func(_ context.Context, userID string, reqType domain.RequestType, reason string) (*domain.DataRequest, error) {
			return sampleRequest(), nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodPost, "/gdpr/requests", map[string]string{
		"user_id": "user-1",
		"type":    "export",
		"reason":  "I want my data",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var dr domain.DataRequest
	if err := json.NewDecoder(rec.Body).Decode(&dr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dr.ID != "req-1" {
		t.Fatalf("want id=req-1, got %s", dr.ID)
	}
}

func TestSubmitRequest_MissingFields(t *testing.T) {
	h := newHandler(&mockService{})
	rec := do(t, h, http.MethodPost, "/gdpr/requests", map[string]string{"reason": "oops"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestGetRequest_Found(t *testing.T) {
	svc := &mockService{
		getRequestFn: func(_ context.Context, id string) (*domain.DataRequest, error) {
			if id != "req-1" {
				return nil, domain.ErrNotFound
			}
			return sampleRequest(), nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodGet, "/gdpr/requests/req-1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetRequest_NotFound(t *testing.T) {
	svc := &mockService{
		getRequestFn: func(_ context.Context, id string) (*domain.DataRequest, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodGet, "/gdpr/requests/does-not-exist", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestListRequests(t *testing.T) {
	svc := &mockService{
		listRequestsFn: func(_ context.Context, userID string) ([]*domain.DataRequest, error) {
			return []*domain.DataRequest{sampleRequest()}, nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodGet, "/gdpr/users/user-1/requests", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []*domain.DataRequest
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 item, got %d", len(list))
	}
}

func TestProcessRequest_NoContent(t *testing.T) {
	svc := &mockService{
		processRequestFn: func(_ context.Context, id string, notes string) error {
			return nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodPost, "/gdpr/requests/req-1/process", map[string]string{"notes": "starting"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProcessRequest_NotFound(t *testing.T) {
	svc := &mockService{
		processRequestFn: func(_ context.Context, id string, notes string) error {
			return domain.ErrNotFound
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodPost, "/gdpr/requests/missing/process", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestCompleteRequest_NoContent(t *testing.T) {
	svc := &mockService{
		completeRequestFn: func(_ context.Context, id string, notes string) error {
			return nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodPost, "/gdpr/requests/req-1/complete", map[string]string{"notes": "done"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRejectRequest_NoContent(t *testing.T) {
	svc := &mockService{
		rejectRequestFn: func(_ context.Context, id string, notes string) error {
			return nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodPost, "/gdpr/requests/req-1/reject", map[string]string{"notes": "invalid"})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateConsent_OK(t *testing.T) {
	svc := &mockService{
		updateConsentFn: func(_ context.Context, userID string, consentType domain.ConsentType, granted bool, ip string) error {
			return nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodPut, "/gdpr/users/user-1/consent", map[string]interface{}{
		"type":       "marketing",
		"granted":    true,
		"ip_address": "10.0.0.1",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateConsent_MissingType(t *testing.T) {
	h := newHandler(&mockService{})
	rec := do(t, h, http.MethodPut, "/gdpr/users/user-1/consent", map[string]interface{}{
		"granted": true,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestGetConsents_OK(t *testing.T) {
	svc := &mockService{
		getConsentsFn: func(_ context.Context, userID string) ([]*domain.Consent, error) {
			return []*domain.Consent{
				{UserID: userID, Type: domain.ConsentMarketing, Granted: true, IPAddress: "10.0.0.1"},
			}, nil
		},
	}
	h := newHandler(svc)
	rec := do(t, h, http.MethodGet, "/gdpr/users/user-1/consent", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var list []*domain.Consent
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 consent, got %d", len(list))
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newHandler(&mockService{})
	// DELETE is not wired
	rec := do(t, h, http.MethodDelete, "/gdpr/requests", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", rec.Code)
	}
}

func TestUnknownRoute(t *testing.T) {
	h := newHandler(&mockService{})
	rec := do(t, h, http.MethodGet, "/gdpr/requests/req-1/unknown-action", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}
