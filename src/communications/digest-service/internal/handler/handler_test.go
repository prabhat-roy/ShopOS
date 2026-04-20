package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopos/digest-service/internal/domain"
	"github.com/shopos/digest-service/internal/handler"
)

// mockStore satisfies store.Storer for unit tests.
type mockStore struct {
	createFn      func(ctx context.Context, cfg domain.DigestConfig) error
	getFn         func(ctx context.Context, id uuid.UUID) (domain.DigestConfig, error)
	getByUserFn   func(ctx context.Context, userID uuid.UUID) ([]domain.DigestConfig, error)
	listFn        func(ctx context.Context, status domain.DigestStatus, frequency domain.DigestFrequency) ([]domain.DigestConfig, error)
	updateNextFn  func(ctx context.Context, id uuid.UUID, t time.Time) error
	updateLastFn  func(ctx context.Context, id uuid.UUID, t time.Time) error
	pauseFn       func(ctx context.Context, id uuid.UUID) error
	resumeFn      func(ctx context.Context, id uuid.UUID) error
	deleteFn      func(ctx context.Context, id uuid.UUID) error
	saveRunFn     func(ctx context.Context, run domain.DigestRun) error
	listRunsFn    func(ctx context.Context, configID uuid.UUID, limit int) ([]domain.DigestRun, error)
	listDueFn     func(ctx context.Context, now time.Time) ([]domain.DigestConfig, error)
}

func (m *mockStore) CreateConfig(ctx context.Context, cfg domain.DigestConfig) error {
	return m.createFn(ctx, cfg)
}
func (m *mockStore) GetConfig(ctx context.Context, id uuid.UUID) (domain.DigestConfig, error) {
	return m.getFn(ctx, id)
}
func (m *mockStore) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.DigestConfig, error) {
	return m.getByUserFn(ctx, userID)
}
func (m *mockStore) ListConfigs(ctx context.Context, status domain.DigestStatus, frequency domain.DigestFrequency) ([]domain.DigestConfig, error) {
	return m.listFn(ctx, status, frequency)
}
func (m *mockStore) UpdateNextSend(ctx context.Context, id uuid.UUID, t time.Time) error {
	return m.updateNextFn(ctx, id, t)
}
func (m *mockStore) UpdateLastSent(ctx context.Context, id uuid.UUID, t time.Time) error {
	return m.updateLastFn(ctx, id, t)
}
func (m *mockStore) PauseConfig(ctx context.Context, id uuid.UUID) error {
	return m.pauseFn(ctx, id)
}
func (m *mockStore) ResumeConfig(ctx context.Context, id uuid.UUID) error {
	return m.resumeFn(ctx, id)
}
func (m *mockStore) DeleteConfig(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}
func (m *mockStore) SaveRun(ctx context.Context, run domain.DigestRun) error {
	return m.saveRunFn(ctx, run)
}
func (m *mockStore) ListRuns(ctx context.Context, configID uuid.UUID, limit int) ([]domain.DigestRun, error) {
	return m.listRunsFn(ctx, configID, limit)
}
func (m *mockStore) ListDueConfigs(ctx context.Context, now time.Time) ([]domain.DigestConfig, error) {
	return m.listDueFn(ctx, now)
}

func newTestServer(s *mockStore) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(s)
	h.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func sampleConfig() domain.DigestConfig {
	return domain.DigestConfig{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Email:      "user@example.com",
		Frequency:  domain.FrequencyDaily,
		Status:     domain.StatusActive,
		NextSendAt: time.Now().Add(24 * time.Hour),
		Timezone:   "UTC",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Test 1: GET /healthz returns 200.
func TestHealthz(t *testing.T) {
	srv := newTestServer(&mockStore{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 2: POST /digests with valid payload returns 201.
func TestCreateConfig_Created(t *testing.T) {
	s := &mockStore{
		createFn: func(_ context.Context, _ domain.DigestConfig) error { return nil },
	}
	srv := newTestServer(s)
	defer srv.Close()

	payload := fmt.Sprintf(`{"userId":%q,"email":"u@example.com","frequency":"DAILY","timezone":"UTC"}`, uuid.New().String())
	resp, err := http.Post(srv.URL+"/digests", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

// Test 3: POST /digests with missing email returns 400.
func TestCreateConfig_BadRequest(t *testing.T) {
	srv := newTestServer(&mockStore{})
	defer srv.Close()

	payload := fmt.Sprintf(`{"userId":%q,"frequency":"DAILY"}`, uuid.New().String())
	resp, err := http.Post(srv.URL+"/digests", "application/json", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 4: GET /digests/{id} returns 200 for a known config.
func TestGetConfig_OK(t *testing.T) {
	cfg := sampleConfig()
	s := &mockStore{
		getFn: func(_ context.Context, _ uuid.UUID) (domain.DigestConfig, error) { return cfg, nil },
	}
	srv := newTestServer(s)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/digests/" + cfg.ID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var got domain.DigestConfig
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != cfg.ID {
		t.Errorf("expected id=%s, got %s", cfg.ID, got.ID)
	}
}

// Test 5: GET /digests/{id} with invalid UUID returns 400.
func TestGetConfig_InvalidID(t *testing.T) {
	srv := newTestServer(&mockStore{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/digests/not-a-uuid")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// Test 6: GET /digests/user/{userId} returns list.
func TestGetByUser(t *testing.T) {
	cfg := sampleConfig()
	s := &mockStore{
		getByUserFn: func(_ context.Context, _ uuid.UUID) ([]domain.DigestConfig, error) {
			return []domain.DigestConfig{cfg}, nil
		},
	}
	srv := newTestServer(s)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/digests/user/" + cfg.UserID.String())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var got []domain.DigestConfig
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 config, got %d", len(got))
	}
}

// Test 7: POST /digests/{id}/pause returns 204.
func TestPauseConfig(t *testing.T) {
	s := &mockStore{
		pauseFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	srv := newTestServer(s)
	defer srv.Close()

	id := uuid.New()
	resp, err := http.Post(srv.URL+"/digests/"+id.String()+"/pause", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 8: POST /digests/{id}/resume returns 204.
func TestResumeConfig(t *testing.T) {
	s := &mockStore{
		resumeFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	srv := newTestServer(s)
	defer srv.Close()

	id := uuid.New()
	resp, err := http.Post(srv.URL+"/digests/"+id.String()+"/resume", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 9: DELETE /digests/{id} returns 204.
func TestDeleteConfig(t *testing.T) {
	s := &mockStore{
		deleteFn: func(_ context.Context, _ uuid.UUID) error { return nil },
	}
	srv := newTestServer(s)
	defer srv.Close()

	id := uuid.New()
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/digests/"+id.String(), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 10: GET /digests/{id}/runs returns list of runs.
func TestListRuns(t *testing.T) {
	run := domain.DigestRun{
		ID:        uuid.New(),
		ConfigID:  uuid.New(),
		SentAt:    time.Now(),
		ItemCount: 3,
		Status:    "SUCCESS",
	}
	s := &mockStore{
		listRunsFn: func(_ context.Context, _ uuid.UUID, _ int) ([]domain.DigestRun, error) {
			return []domain.DigestRun{run}, nil
		},
	}
	srv := newTestServer(s)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/digests/" + run.ConfigID.String() + "/runs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var got []domain.DigestRun
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 run, got %d", len(got))
	}
}
