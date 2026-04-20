package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/shopos/device-fingerprint-service/internal/domain"
	"github.com/shopos/device-fingerprint-service/internal/handler"
	"github.com/shopos/device-fingerprint-service/internal/service"
)

// ---- mock service -----------------------------------------------------------

// mockService is a test double for service.Servicer.
type mockService struct {
	identifyFn            func(ctx context.Context, req *domain.FingerprintRequest) (*domain.FingerprintResponse, error)
	getByIDFn             func(ctx context.Context, id string) (*domain.Fingerprint, error)
	getUserFingerprintsFn func(ctx context.Context, userID string) ([]*domain.Fingerprint, error)
}

func (m *mockService) Identify(ctx context.Context, req *domain.FingerprintRequest) (*domain.FingerprintResponse, error) {
	return m.identifyFn(ctx, req)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*domain.Fingerprint, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockService) GetUserFingerprints(ctx context.Context, userID string) ([]*domain.Fingerprint, error) {
	return m.getUserFingerprintsFn(ctx, userID)
}

var _ service.Servicer = (*mockService)(nil)

// ---- test helpers -----------------------------------------------------------

func newTestHandler(svc service.Servicer) http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	return handler.New(svc, logger)
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	return b
}

// ---- tests ------------------------------------------------------------------

func TestHealthz(t *testing.T) {
	h := newTestHandler(&mockService{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %q", body["status"])
	}
}

func TestIdentify_Success(t *testing.T) {
	want := &domain.FingerprintResponse{
		FingerprintID: "test-uuid-1234",
		Hash:          "abc123hash",
		IsKnown:       false,
		TrustScore:    10,
	}

	svc := &mockService{
		identifyFn: func(_ context.Context, _ *domain.FingerprintRequest) (*domain.FingerprintResponse, error) {
			return want, nil
		},
	}

	h := newTestHandler(svc)

	reqBody := domain.FingerprintRequest{
		UserID: "user-42",
		Attributes: domain.DeviceAttributes{
			UserAgent:  "Mozilla/5.0",
			AcceptLang: "en-US",
			Timezone:   "UTC",
			ScreenRes:  "1920x1080",
			Platform:   "Linux x86_64",
			IPAddress:  "127.0.0.1",
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/fingerprints/identify", bytes.NewReader(mustMarshal(t, reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var got domain.FingerprintResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.FingerprintID != want.FingerprintID {
		t.Errorf("FingerprintID: got %q, want %q", got.FingerprintID, want.FingerprintID)
	}
	if got.Hash != want.Hash {
		t.Errorf("Hash: got %q, want %q", got.Hash, want.Hash)
	}
	if got.IsKnown != want.IsKnown {
		t.Errorf("IsKnown: got %v, want %v", got.IsKnown, want.IsKnown)
	}
	if got.TrustScore != want.TrustScore {
		t.Errorf("TrustScore: got %d, want %d", got.TrustScore, want.TrustScore)
	}
}

func TestIdentify_MissingUserAgent(t *testing.T) {
	h := newTestHandler(&mockService{})

	reqBody := domain.FingerprintRequest{
		UserID: "user-42",
		Attributes: domain.DeviceAttributes{
			AcceptLang: "en-US",
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/fingerprints/identify", bytes.NewReader(mustMarshal(t, reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIdentify_BadJSON(t *testing.T) {
	h := newTestHandler(&mockService{})

	req := httptest.NewRequest(http.MethodPost, "/fingerprints/identify", bytes.NewReader([]byte(`{bad json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetByID_Found(t *testing.T) {
	wantFP := &domain.Fingerprint{
		ID:         "fp-uuid-999",
		Hash:       "deadbeef",
		UserID:     "user-42",
		TrustScore: 100,
		SeenCount:  10,
	}

	svc := &mockService{
		getByIDFn: func(_ context.Context, id string) (*domain.Fingerprint, error) {
			if id == "fp-uuid-999" {
				return wantFP, nil
			}
			return nil, domain.ErrNotFound
		},
	}

	h := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/fingerprints/fp-uuid-999", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var got domain.Fingerprint
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.ID != wantFP.ID {
		t.Errorf("ID: got %q, want %q", got.ID, wantFP.ID)
	}
	if got.Hash != wantFP.Hash {
		t.Errorf("Hash: got %q, want %q", got.Hash, wantFP.Hash)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(_ context.Context, _ string) (*domain.Fingerprint, error) {
			return nil, domain.ErrNotFound
		},
	}

	h := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/fingerprints/does-not-exist", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetUserFingerprints_Success(t *testing.T) {
	fps := []*domain.Fingerprint{
		{ID: "fp-1", Hash: "aaa", UserID: "user-7"},
		{ID: "fp-2", Hash: "bbb", UserID: "user-7"},
	}

	svc := &mockService{
		getUserFingerprintsFn: func(_ context.Context, userID string) ([]*domain.Fingerprint, error) {
			if userID == "user-7" {
				return fps, nil
			}
			return []*domain.Fingerprint{}, nil
		},
	}

	h := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/fingerprints/user/user-7", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var got []*domain.Fingerprint
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 fingerprints, got %d", len(got))
	}
	if got[0].ID != "fp-1" {
		t.Errorf("first fp ID: got %q, want fp-1", got[0].ID)
	}
}

func TestGetUserFingerprints_Empty(t *testing.T) {
	svc := &mockService{
		getUserFingerprintsFn: func(_ context.Context, _ string) ([]*domain.Fingerprint, error) {
			return []*domain.Fingerprint{}, nil
		},
	}

	h := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/fingerprints/user/nobody", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got []*domain.Fingerprint
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}
