package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shopos/mfa-service/internal/domain"
	"github.com/shopos/mfa-service/internal/handler"
)

// mockService implements handler.Servicer for unit tests.
type mockService struct {
	enrollFn    func(ctx context.Context, userID string) (*domain.MFASetup, error)
	confirmFn   func(ctx context.Context, userID, code string) error
	verifyFn    func(ctx context.Context, req *domain.VerifyRequest) (*domain.VerifyResponse, error)
	disableFn   func(ctx context.Context, userID string) error
	getStatusFn func(ctx context.Context, userID string) (domain.MFAStatus, error)
}

func (m *mockService) Enroll(ctx context.Context, userID string) (*domain.MFASetup, error) {
	return m.enrollFn(ctx, userID)
}
func (m *mockService) Confirm(ctx context.Context, userID, code string) error {
	return m.confirmFn(ctx, userID, code)
}
func (m *mockService) Verify(ctx context.Context, req *domain.VerifyRequest) (*domain.VerifyResponse, error) {
	return m.verifyFn(ctx, req)
}
func (m *mockService) Disable(ctx context.Context, userID string) error {
	return m.disableFn(ctx, userID)
}
func (m *mockService) GetStatus(ctx context.Context, userID string) (domain.MFAStatus, error) {
	return m.getStatusFn(ctx, userID)
}

// newHandler returns an http.Handler backed by the provided mock.
func newHandler(svc handler.Servicer) http.Handler {
	logger := log.New(os.Stderr, "[test] ", 0)
	return handler.New(svc, logger)
}

// doRequest performs a test HTTP request and returns the recorder.
func doRequest(h http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// ---- /healthz ----------------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := newHandler(&mockService{})
	rec := doRequest(h, http.MethodGet, "/healthz", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", resp["status"])
	}
}

func TestHealthzMethodNotAllowed(t *testing.T) {
	h := newHandler(&mockService{})
	rec := doRequest(h, http.MethodPost, "/healthz", nil)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

// ---- POST /mfa/{userID}/enroll -----------------------------------------------

func TestEnroll_Success(t *testing.T) {
	svc := &mockService{
		enrollFn: func(_ context.Context, userID string) (*domain.MFASetup, error) {
			return &domain.MFASetup{
				UserID:      userID,
				Secret:      "SECRETBASE32",
				QRCodeURL:   "otpauth://totp/ShopOS:user1?secret=SECRETBASE32",
				BackupCodes: []string{"aabbccdd", "eeff0011"},
				Status:      domain.MFAPending,
			}, nil
		},
	}

	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/enroll", nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["user_id"] != "user1" {
		t.Errorf("unexpected user_id: %v", resp["user_id"])
	}
	if resp["qr_code_url"] == "" {
		t.Error("expected qr_code_url to be non-empty")
	}
	codes, ok := resp["backup_codes"].([]interface{})
	if !ok || len(codes) == 0 {
		t.Error("expected backup_codes to be a non-empty array")
	}
}

func TestEnroll_AlreadyEnabled(t *testing.T) {
	svc := &mockService{
		enrollFn: func(_ context.Context, _ string) (*domain.MFASetup, error) {
			return nil, domain.ErrAlreadyEnabled
		},
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/enroll", nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

// ---- POST /mfa/{userID}/confirm ----------------------------------------------

func TestConfirm_Success(t *testing.T) {
	svc := &mockService{
		confirmFn: func(_ context.Context, _, _ string) error { return nil },
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/confirm",
		map[string]string{"code": "123456"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConfirm_InvalidCode(t *testing.T) {
	svc := &mockService{
		confirmFn: func(_ context.Context, _, _ string) error { return domain.ErrInvalidCode },
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/confirm",
		map[string]string{"code": "000000"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConfirm_NotConfigured(t *testing.T) {
	svc := &mockService{
		confirmFn: func(_ context.Context, _, _ string) error { return domain.ErrNotFound },
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/confirm",
		map[string]string{"code": "123456"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestConfirm_MissingCode(t *testing.T) {
	svc := &mockService{}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/confirm",
		map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---- POST /mfa/{userID}/verify -----------------------------------------------

func TestVerify_SuccessTOTP(t *testing.T) {
	svc := &mockService{
		verifyFn: func(_ context.Context, _ *domain.VerifyRequest) (*domain.VerifyResponse, error) {
			return &domain.VerifyResponse{Valid: true, IsBackupCode: false}, nil
		},
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/verify",
		map[string]string{"code": "123456"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["valid"] != true {
		t.Errorf("expected valid=true")
	}
	if resp["is_backup_code"] != false {
		t.Errorf("expected is_backup_code=false")
	}
}

func TestVerify_SuccessBackupCode(t *testing.T) {
	svc := &mockService{
		verifyFn: func(_ context.Context, _ *domain.VerifyRequest) (*domain.VerifyResponse, error) {
			return &domain.VerifyResponse{Valid: true, IsBackupCode: true}, nil
		},
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/verify",
		map[string]string{"code": "aabbccdd"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["is_backup_code"] != true {
		t.Errorf("expected is_backup_code=true")
	}
}

func TestVerify_InvalidCode(t *testing.T) {
	svc := &mockService{
		verifyFn: func(_ context.Context, _ *domain.VerifyRequest) (*domain.VerifyResponse, error) {
			return nil, domain.ErrInvalidCode
		},
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/verify",
		map[string]string{"code": "badcode"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestVerify_NotConfigured(t *testing.T) {
	svc := &mockService{
		verifyFn: func(_ context.Context, _ *domain.VerifyRequest) (*domain.VerifyResponse, error) {
			return nil, domain.ErrNotFound
		},
	}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/verify",
		map[string]string{"code": "123456"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestVerify_MissingCode(t *testing.T) {
	svc := &mockService{}
	rec := doRequest(newHandler(svc), http.MethodPost, "/mfa/user1/verify",
		map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ---- DELETE /mfa/{userID} ----------------------------------------------------

func TestDisable_Success(t *testing.T) {
	svc := &mockService{
		disableFn: func(_ context.Context, _ string) error { return nil },
	}
	rec := doRequest(newHandler(svc), http.MethodDelete, "/mfa/user1", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestDisable_NotConfigured(t *testing.T) {
	svc := &mockService{
		disableFn: func(_ context.Context, _ string) error { return domain.ErrNotFound },
	}
	rec := doRequest(newHandler(svc), http.MethodDelete, "/mfa/user1", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---- GET /mfa/{userID}/status ------------------------------------------------

func TestGetStatus_Enabled(t *testing.T) {
	svc := &mockService{
		getStatusFn: func(_ context.Context, _ string) (domain.MFAStatus, error) {
			return domain.MFAEnabled, nil
		},
	}
	rec := doRequest(newHandler(svc), http.MethodGet, "/mfa/user1/status", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "enabled" {
		t.Errorf("expected status=enabled, got %q", resp["status"])
	}
	if resp["user_id"] != "user1" {
		t.Errorf("expected user_id=user1, got %q", resp["user_id"])
	}
}

func TestGetStatus_NotConfigured(t *testing.T) {
	svc := &mockService{
		getStatusFn: func(_ context.Context, _ string) (domain.MFAStatus, error) {
			return "", domain.ErrNotFound
		},
	}
	rec := doRequest(newHandler(svc), http.MethodGet, "/mfa/user1/status", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ---- Unknown routes ----------------------------------------------------------

func TestUnknownRoute(t *testing.T) {
	h := newHandler(&mockService{})
	rec := doRequest(h, http.MethodGet, "/mfa/user1/unknown", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestMFARouteNoUserID(t *testing.T) {
	h := newHandler(&mockService{})
	// "/mfa/" with nothing after the slash
	rec := doRequest(h, http.MethodGet, "/mfa/", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
