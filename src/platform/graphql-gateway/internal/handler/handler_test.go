package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enterprise/graphql-gateway/internal/handler"
)

// -------------------------------------------------------------------------
// Mock Executer
// -------------------------------------------------------------------------

type mockExecuter struct {
	data   map[string]any
	errors []string
}

func (m *mockExecuter) Execute(_ context.Context, _ string) (map[string]any, []string) {
	return m.data, m.errors
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

func newServer(exec handler.Executer) *httptest.Server {
	mux := http.NewServeMux()
	h := handler.New(exec)
	h.Register(mux)
	return httptest.NewServer(mux)
}

// -------------------------------------------------------------------------
// POST /graphql tests
// -------------------------------------------------------------------------

// TestGraphQL_ValidQuery verifies that a well-formed POST /graphql request
// returns HTTP 200 and a body containing a "data" field.
func TestGraphQL_ValidQuery(t *testing.T) {
	exec := &mockExecuter{
		data: map[string]any{
			"products": []any{
				map[string]any{"id": "p1", "name": "Widget"},
			},
		},
		errors: nil,
	}

	srv := newServer(exec)
	defer srv.Close()

	body := `{"query":"{ products(limit:1) }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /graphql failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := result["data"]; !ok {
		t.Error("response should contain 'data' key")
	}
}

// TestGraphQL_WithFieldErrors verifies that field-level errors are returned in
// the "errors" array while "data" is still present.
func TestGraphQL_WithFieldErrors(t *testing.T) {
	exec := &mockExecuter{
		data:   map[string]any{"categories": []any{}},
		errors: []string{`field "unknownField": unknown field`},
	}

	srv := newServer(exec)
	defer srv.Close()

	body := `{"query":"{ categories unknownField }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /graphql failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 even with field errors, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := result["errors"]; !ok {
		t.Error("response should contain 'errors' key when field errors exist")
	}
	if _, ok := result["data"]; !ok {
		t.Error("response should still contain 'data' key alongside errors")
	}
}

// TestGraphQL_EmptyBody verifies that a POST /graphql with no body returns 400.
func TestGraphQL_EmptyBody(t *testing.T) {
	exec := &mockExecuter{}
	srv := newServer(exec)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/graphql", "application/json", bytes.NewReader([]byte{}))
	if err != nil {
		t.Fatalf("POST /graphql failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", resp.StatusCode)
	}
}

// TestGraphQL_MissingQueryField verifies that a POST /graphql with a JSON body
// that omits the "query" field returns 400.
func TestGraphQL_MissingQueryField(t *testing.T) {
	exec := &mockExecuter{}
	srv := newServer(exec)
	defer srv.Close()

	body := `{"mutation":"{ createProduct }"}`
	resp, err := http.Post(srv.URL+"/graphql", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /graphql failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 when 'query' is missing, got %d", resp.StatusCode)
	}
}

// TestGraphQL_WrongMethod verifies that GET /graphql returns 405.
func TestGraphQL_WrongMethod(t *testing.T) {
	exec := &mockExecuter{}
	srv := newServer(exec)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/graphql")
	if err != nil {
		t.Fatalf("GET /graphql failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET /graphql, got %d", resp.StatusCode)
	}
}

// -------------------------------------------------------------------------
// GET /healthz tests
// -------------------------------------------------------------------------

// TestHealthz_OK verifies that GET /healthz returns HTTP 200 and {"status":"ok"}.
func TestHealthz_OK(t *testing.T) {
	exec := &mockExecuter{}
	srv := newServer(exec)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode /healthz response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", result["status"])
	}
}

// TestHealthz_WrongMethod verifies that POST /healthz returns 405.
func TestHealthz_WrongMethod(t *testing.T) {
	exec := &mockExecuter{}
	srv := newServer(exec)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/healthz", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /healthz failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for POST /healthz, got %d", resp.StatusCode)
	}
}

// -------------------------------------------------------------------------
// GET /schema tests
// -------------------------------------------------------------------------

// TestSchema_OK verifies that GET /schema returns HTTP 200 and a JSON body
// containing a "fields" array.
func TestSchema_OK(t *testing.T) {
	exec := &mockExecuter{}
	srv := newServer(exec)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/schema")
	if err != nil {
		t.Fatalf("GET /schema failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode /schema response: %v", err)
	}
	fields, ok := result["fields"]
	if !ok {
		t.Fatal("expected 'fields' key in schema response")
	}
	list, ok := fields.([]any)
	if !ok {
		t.Fatalf("expected 'fields' to be an array, got %T", fields)
	}
	if len(list) == 0 {
		t.Error("schema should describe at least one field")
	}
}
