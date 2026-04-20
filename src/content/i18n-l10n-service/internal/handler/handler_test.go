package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopos/i18n-l10n-service/internal/domain"
	"github.com/shopos/i18n-l10n-service/internal/handler"
)

// ─── Mock service ─────────────────────────────────────────────────────────────

type mockService struct {
	getTranslationFn  func(locale, namespace, key string) (string, error)
	getNamespaceFn    func(locale, namespace string) (map[string]string, error)
	upsertFn          func(t domain.Translation) error
	bulkUpsertFn      func(req domain.BulkUpsertRequest) error
	deleteFn          func(locale, namespace, key string) error
	listLocalesFn     func() ([]string, error)
	listNamespacesFn  func(locale string) ([]string, error)
}

func (m *mockService) GetTranslation(locale, namespace, key string) (string, error) {
	if m.getTranslationFn != nil {
		return m.getTranslationFn(locale, namespace, key)
	}
	return "", domain.ErrNotFound
}
func (m *mockService) GetNamespace(locale, namespace string) (map[string]string, error) {
	if m.getNamespaceFn != nil {
		return m.getNamespaceFn(locale, namespace)
	}
	return nil, nil
}
func (m *mockService) UpsertTranslation(t domain.Translation) error {
	if m.upsertFn != nil {
		return m.upsertFn(t)
	}
	return nil
}
func (m *mockService) BulkUpsert(req domain.BulkUpsertRequest) error {
	if m.bulkUpsertFn != nil {
		return m.bulkUpsertFn(req)
	}
	return nil
}
func (m *mockService) DeleteTranslation(locale, namespace, key string) error {
	if m.deleteFn != nil {
		return m.deleteFn(locale, namespace, key)
	}
	return nil
}
func (m *mockService) ListLocales() ([]string, error) {
	if m.listLocalesFn != nil {
		return m.listLocalesFn()
	}
	return []string{"en"}, nil
}
func (m *mockService) ListNamespaces(locale string) ([]string, error) {
	if m.listNamespacesFn != nil {
		return m.listNamespacesFn(locale)
	}
	return []string{"common"}, nil
}
func (m *mockService) GetSupportedLocales() []string {
	return domain.SupportedLocales
}

// ─── Helper ───────────────────────────────────────────────────────────────────

func newTestServer(svc *mockService) *httptest.Server {
	h := handler.New(svc)
	return httptest.NewServer(h)
}

func jsonBody(v any) *bytes.Buffer {
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

// ─── Tests ────────────────────────────────────────────────────────────────────

// Test 1: GET /healthz
func TestHealthz(t *testing.T) {
	srv := newTestServer(&mockService{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 2: GET /i18n/{locale}/{namespace}/{key} — found
func TestGetTranslation_Found(t *testing.T) {
	svc := &mockService{
		getTranslationFn: func(locale, namespace, key string) (string, error) {
			if locale == "en" && namespace == "common" && key == "hello" {
				return "Hello", nil
			}
			return "", domain.ErrNotFound
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/i18n/en/common/hello")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	if body["value"] != "Hello" {
		t.Fatalf("expected value=Hello, got %q", body["value"])
	}
}

// Test 3: GET /i18n/{locale}/{namespace}/{key} — not found
func TestGetTranslation_NotFound(t *testing.T) {
	svc := &mockService{
		getTranslationFn: func(_, _, _ string) (string, error) {
			return "", domain.ErrNotFound
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/i18n/fr/missing/key")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 4: GET /i18n/{locale}/{namespace} — whole namespace
func TestGetNamespace(t *testing.T) {
	svc := &mockService{
		getNamespaceFn: func(locale, namespace string) (map[string]string, error) {
			return map[string]string{"hello": "Hola", "bye": "Adiós"}, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/i18n/es/common")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	translations := body["translations"].(map[string]any)
	if translations["hello"] != "Hola" {
		t.Fatalf("expected Hola, got %v", translations["hello"])
	}
}

// Test 5: PUT /i18n/{locale}/{namespace}/{key} — create (201)
func TestUpsertTranslation_Create(t *testing.T) {
	callCount := 0
	svc := &mockService{
		getTranslationFn: func(_, _, _ string) (string, error) {
			return "", domain.ErrNotFound // simulate new key
		},
		upsertFn: func(t domain.Translation) error {
			callCount++
			return nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut,
		srv.URL+"/i18n/en/common/new_key",
		jsonBody(map[string]string{"value": "New Value"}),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if callCount != 1 {
		t.Fatalf("expected upsert called once, got %d", callCount)
	}
}

// Test 6: PUT /i18n/{locale}/{namespace}/{key} — update (200)
func TestUpsertTranslation_Update(t *testing.T) {
	svc := &mockService{
		getTranslationFn: func(_, _, _ string) (string, error) {
			return "existing", nil // exists → 200
		},
		upsertFn: func(_ domain.Translation) error { return nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPut,
		srv.URL+"/i18n/en/common/existing_key",
		jsonBody(map[string]string{"value": "Updated"}),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// Test 7: PUT /i18n/{locale}/{namespace} — bulk upsert
func TestBulkUpsert(t *testing.T) {
	var receivedReq domain.BulkUpsertRequest
	svc := &mockService{
		bulkUpsertFn: func(req domain.BulkUpsertRequest) error {
			receivedReq = req
			return nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	payload := map[string]string{"k1": "v1", "k2": "v2"}
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/i18n/de/common", jsonBody(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if receivedReq.Locale != "de" {
		t.Fatalf("expected locale=de, got %q", receivedReq.Locale)
	}
	if len(receivedReq.Translations) != 2 {
		t.Fatalf("expected 2 translations, got %d", len(receivedReq.Translations))
	}
}

// Test 8: DELETE /i18n/{locale}/{namespace}/{key} — success 204
func TestDeleteTranslation_Success(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_, _, _ string) error { return nil },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/i18n/en/common/bye", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

// Test 9: DELETE /i18n/{locale}/{namespace}/{key} — not found 404
func TestDeleteTranslation_NotFound(t *testing.T) {
	svc := &mockService{
		deleteFn: func(_, _, _ string) error { return domain.ErrNotFound },
	}
	srv := newTestServer(svc)
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/i18n/en/ns/ghost", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

// Test 10: GET /i18n/locales
func TestListLocales(t *testing.T) {
	svc := &mockService{
		listLocalesFn: func() ([]string, error) {
			return []string{"en", "fr", "de"}, nil
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/i18n/locales")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]any
	json.NewDecoder(resp.Body).Decode(&body)
	locales := body["locales"].([]any)
	if len(locales) != 3 {
		t.Fatalf("expected 3 locales, got %d", len(locales))
	}
}

// Extra: ensures service error propagates as 500
func TestGetTranslation_ServiceError(t *testing.T) {
	svc := &mockService{
		getTranslationFn: func(_, _, _ string) (string, error) {
			return "", errors.New("db connection failed")
		},
	}
	srv := newTestServer(svc)
	defer srv.Close()

	resp, _ := http.Get(fmt.Sprintf("%s/i18n/en/common/key", srv.URL))
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}
