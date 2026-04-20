package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/geolocation-service/internal/domain"
)

// mockLooker implements Looker for testing without a real lookup layer.
type mockLooker struct {
	resolveFunc     func(ip string) (*domain.Location, error)
	resolveManyFunc func(ips []string) ([]*domain.Location, error)
}

func (m *mockLooker) Resolve(ip string) (*domain.Location, error) {
	return m.resolveFunc(ip)
}

func (m *mockLooker) ResolveMany(ips []string) ([]*domain.Location, error) {
	return m.resolveManyFunc(ips)
}

// fixedLocation returns a predictable location used across multiple tests.
func fixedLocation(ip string) *domain.Location {
	return &domain.Location{
		IP:          ip,
		CountryCode: "US",
		CountryName: "United States",
		Region:      "California",
		City:        "Mountain View",
		PostalCode:  "94043",
		Timezone:    "America/Los_Angeles",
		ISP:         "Google LLC",
		Latitude:    37.3861,
		Longitude:   -122.0839,
	}
}

// ── /healthz ────────────────────────────────────────────────────────────────

func TestHealthz_OK(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

// ── GET /locate?ip={ip} ─────────────────────────────────────────────────────

func TestLocateQuery_OK(t *testing.T) {
	looker := &mockLooker{
		resolveFunc: func(ip string) (*domain.Location, error) {
			return fixedLocation(ip), nil
		},
	}
	h := New(looker)

	req := httptest.NewRequest(http.MethodGet, "/locate?ip=8.8.8.8", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var loc domain.Location
	if err := json.NewDecoder(w.Body).Decode(&loc); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if loc.CountryCode != "US" {
		t.Errorf("expected US, got %q", loc.CountryCode)
	}
}

func TestLocateQuery_MissingIP(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodGet, "/locate", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLocateQuery_InvalidIP(t *testing.T) {
	looker := &mockLooker{
		resolveFunc: func(ip string) (*domain.Location, error) {
			return nil, domain.ErrInvalidIP
		},
	}
	h := New(looker)

	req := httptest.NewRequest(http.MethodGet, "/locate?ip=not-an-ip", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLocateQuery_NotFound(t *testing.T) {
	looker := &mockLooker{
		resolveFunc: func(ip string) (*domain.Location, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := New(looker)

	req := httptest.NewRequest(http.MethodGet, "/locate?ip=10.0.0.1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ── GET /locate/{ip} ────────────────────────────────────────────────────────

func TestLocatePath_OK(t *testing.T) {
	looker := &mockLooker{
		resolveFunc: func(ip string) (*domain.Location, error) {
			return fixedLocation(ip), nil
		},
	}
	h := New(looker)

	req := httptest.NewRequest(http.MethodGet, "/locate/8.8.8.8", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var loc domain.Location
	if err := json.NewDecoder(w.Body).Decode(&loc); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if loc.City != "Mountain View" {
		t.Errorf("expected Mountain View, got %q", loc.City)
	}
}

func TestLocatePath_InvalidIP(t *testing.T) {
	looker := &mockLooker{
		resolveFunc: func(ip string) (*domain.Location, error) {
			return nil, domain.ErrInvalidIP
		},
	}
	h := New(looker)

	req := httptest.NewRequest(http.MethodGet, "/locate/not-an-ip", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLocatePath_NotFound(t *testing.T) {
	looker := &mockLooker{
		resolveFunc: func(ip string) (*domain.Location, error) {
			return nil, domain.ErrNotFound
		},
	}
	h := New(looker)

	req := httptest.NewRequest(http.MethodGet, "/locate/10.0.0.1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestLocatePath_MethodNotAllowed(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodDelete, "/locate/8.8.8.8", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

// ── POST /locate/batch ──────────────────────────────────────────────────────

func TestBatch_OK(t *testing.T) {
	looker := &mockLooker{
		resolveManyFunc: func(ips []string) ([]*domain.Location, error) {
			locs := make([]*domain.Location, 0, len(ips))
			for _, ip := range ips {
				locs = append(locs, fixedLocation(ip))
			}
			return locs, nil
		},
	}
	h := New(looker)

	body := `{"ips":["8.8.8.8","64.1.2.3"]}`
	req := httptest.NewRequest(http.MethodPost, "/locate/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	count, ok := resp["count"].(float64)
	if !ok || int(count) != 2 {
		t.Errorf("expected count=2, got %v", resp["count"])
	}
}

func TestBatch_EmptyIPs(t *testing.T) {
	h := New(&mockLooker{})

	body := `{"ips":[]}`
	req := httptest.NewRequest(http.MethodPost, "/locate/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBatch_InvalidJSON(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodPost, "/locate/batch", bytes.NewBufferString("{not-json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestBatch_MethodNotAllowed(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodGet, "/locate/batch", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestBatch_LookupError(t *testing.T) {
	looker := &mockLooker{
		resolveManyFunc: func(ips []string) ([]*domain.Location, error) {
			return nil, errors.New("storage unavailable")
		},
	}
	h := New(looker)

	body := `{"ips":["8.8.8.8"]}`
	req := httptest.NewRequest(http.MethodPost, "/locate/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestBatch_PartialResults(t *testing.T) {
	looker := &mockLooker{
		// Only one result even though two IPs were sent (simulates partial resolve).
		resolveManyFunc: func(ips []string) ([]*domain.Location, error) {
			return []*domain.Location{fixedLocation("8.8.8.8")}, nil
		},
	}
	h := New(looker)

	body := `{"ips":["8.8.8.8","10.0.0.1"]}`
	req := httptest.NewRequest(http.MethodPost, "/locate/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	count, ok := resp["count"].(float64)
	if !ok || int(count) != 1 {
		t.Errorf("expected count=1, got %v", resp["count"])
	}
}

// ── Content-Type header ──────────────────────────────────────────────────────

func TestContentTypeJSON(t *testing.T) {
	h := New(&mockLooker{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
