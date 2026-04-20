package resolver_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/enterprise/graphql-gateway/internal/config"
	"github.com/enterprise/graphql-gateway/internal/resolver"
)

// -------------------------------------------------------------------------
// Mock Doer
// -------------------------------------------------------------------------

// mockDoer returns configurable responses keyed by URL prefix.
type mockDoer struct {
	responses map[string]mockResponse
}

type mockResponse struct {
	body   []byte
	status int
	err    error
}

func (m *mockDoer) Get(_ context.Context, url string) ([]byte, int, error) {
	for prefix, resp := range m.responses {
		if len(url) >= len(prefix) && url[:len(prefix)] == prefix {
			return resp.body, resp.status, resp.err
		}
	}
	// Default: empty JSON array with 200.
	return []byte(`[]`), http.StatusOK, nil
}

func (m *mockDoer) Post(_ context.Context, _ string, _ []byte) ([]byte, int, error) {
	return []byte(`{}`), http.StatusOK, nil
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

func newTestResolver(doer *mockDoer) *resolver.Resolver {
	cfg := &config.Config{
		CatalogURL: "http://catalog",
		CartURL:    "http://cart",
		OrdersURL:  "http://orders",
		UserURL:    "http://user",
	}
	return resolver.New(cfg, doer)
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// -------------------------------------------------------------------------
// Tests
// -------------------------------------------------------------------------

// TestExecute_Products verifies that a query containing "products" returns the
// data under the "products" key in the result map.
func TestExecute_Products(t *testing.T) {
	products := []map[string]any{
		{"id": "p1", "name": "Widget"},
		{"id": "p2", "name": "Gadget"},
	}

	doer := &mockDoer{
		responses: map[string]mockResponse{
			"http://catalog/products": {
				body:   mustJSON(products),
				status: http.StatusOK,
			},
		},
	}

	r := newTestResolver(doer)
	data, errs := r.Execute(context.Background(), `{ products(limit:10,offset:0) }`)

	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}

	raw, ok := data["products"]
	if !ok {
		t.Fatal("expected 'products' key in response data")
	}

	list, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected products to be a slice, got %T", raw)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 products, got %d", len(list))
	}
}

// TestExecute_UnknownField verifies that an unrecognised field name produces a
// field-level error but does not cause the whole request to fail — other valid
// fields in the same query are still resolved.
func TestExecute_UnknownField(t *testing.T) {
	categories := []map[string]any{{"id": "c1", "name": "Electronics"}}

	doer := &mockDoer{
		responses: map[string]mockResponse{
			"http://catalog/categories": {
				body:   mustJSON(categories),
				status: http.StatusOK,
			},
		},
	}

	r := newTestResolver(doer)
	data, errs := r.Execute(context.Background(), `{ categories unknownField }`)

	// There should be exactly one error for the unknown field.
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for unknown field, got %d: %v", len(errs), errs)
	}

	// The valid field should still be present.
	if _, ok := data["categories"]; !ok {
		t.Fatal("expected 'categories' key in response data despite unknown field error")
	}

	// The unknown field should NOT be present in data.
	if _, ok := data["unknownField"]; ok {
		t.Fatal("unknown field should not appear in data map")
	}
}

// TestExecute_MultipleFields verifies that multiple fields in one query are all
// resolved and merged into the result map.
func TestExecute_MultipleFields(t *testing.T) {
	products := []map[string]any{{"id": "p1"}}
	categories := []map[string]any{{"id": "c1"}}

	doer := &mockDoer{
		responses: map[string]mockResponse{
			"http://catalog/products":   {body: mustJSON(products), status: http.StatusOK},
			"http://catalog/categories": {body: mustJSON(categories), status: http.StatusOK},
		},
	}

	r := newTestResolver(doer)
	data, errs := r.Execute(context.Background(), `{ products(limit:5) categories }`)

	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if _, ok := data["products"]; !ok {
		t.Error("missing 'products' in data")
	}
	if _, ok := data["categories"]; !ok {
		t.Error("missing 'categories' in data")
	}
}

// TestExecute_DownstreamError verifies that a downstream HTTP error is captured
// as a field-level error and does not panic.
func TestExecute_DownstreamError(t *testing.T) {
	doer := &mockDoer{
		responses: map[string]mockResponse{
			"http://catalog/products": {
				body:   nil,
				status: http.StatusInternalServerError,
			},
		},
	}

	r := newTestResolver(doer)
	_, errs := r.Execute(context.Background(), `{ products }`)

	if len(errs) == 0 {
		t.Fatal("expected a field-level error for downstream 500, got none")
	}
}

// TestExecute_CartRequiresUserId verifies that calling cart without userId
// produces a field-level error.
func TestExecute_CartRequiresUserId(t *testing.T) {
	doer := &mockDoer{responses: map[string]mockResponse{}}
	r := newTestResolver(doer)

	_, errs := r.Execute(context.Background(), `{ cart }`)
	if len(errs) == 0 {
		t.Fatal("expected error when userId is missing for cart field")
	}
}
