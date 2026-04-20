package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

// Executer is the interface the handler uses to process GraphQL queries.
type Executer interface {
	Execute(ctx context.Context, query string) (map[string]any, []string)
}

// Handler holds a reference to the query executer and exposes HTTP routes.
type Handler struct {
	exec Executer
}

// New creates a new Handler backed by the given Executer.
func New(exec Executer) *Handler {
	return &Handler{exec: exec}
}

// Register wires all routes into the provided ServeMux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/graphql", h.graphqlHandler)
	mux.HandleFunc("/healthz", h.healthzHandler)
	mux.HandleFunc("/schema", h.schemaHandler)
}

// -------------------------------------------------------------------------
// Route handlers
// -------------------------------------------------------------------------

// graphqlHandler handles POST /graphql.
// Expected request body: {"query": "{ ... }"}
// Response body:         {"data":{...},"errors":[...]}
func (h *Handler) graphqlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "request body is required")
		return
	}

	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "field 'query' must not be empty")
		return
	}

	data, fieldErrors := h.exec.Execute(r.Context(), req.Query)

	resp := map[string]any{
		"data": data,
	}
	if len(fieldErrors) > 0 {
		errs := make([]map[string]any, len(fieldErrors))
		for i, msg := range fieldErrors {
			errs[i] = map[string]any{"message": msg}
		}
		resp["errors"] = errs
	}

	writeJSON(w, http.StatusOK, resp)
}

// healthzHandler handles GET /healthz.
func (h *Handler) healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// schemaHandler handles GET /schema and returns a JSON description of all
// available GraphQL-like fields supported by this gateway.
func (h *Handler) schemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	schema := map[string]any{
		"fields": []map[string]any{
			{
				"name":        "products",
				"description": "Paginated list of products from the catalog service.",
				"arguments": []map[string]string{
					{"name": "limit", "type": "Int", "description": "Maximum number of results (default: 20)"},
					{"name": "offset", "type": "Int", "description": "Number of results to skip (default: 0)"},
				},
				"returns": "Array of product objects",
			},
			{
				"name":        "product",
				"description": "Fetch a single product by its ID.",
				"arguments": []map[string]string{
					{"name": "id", "type": "String", "description": "Product ID (required)"},
				},
				"returns": "Product object",
			},
			{
				"name":        "categories",
				"description": "All product categories from the catalog service.",
				"arguments":   []map[string]string{},
				"returns":     "Array of category objects",
			},
			{
				"name":        "cart",
				"description": "Active shopping cart for a given user.",
				"arguments": []map[string]string{
					{"name": "userId", "type": "String", "description": "User ID (required)"},
				},
				"returns": "Cart object",
			},
			{
				"name":        "orders",
				"description": "Paginated list of orders for a given user.",
				"arguments": []map[string]string{
					{"name": "userId", "type": "String", "description": "User ID (required)"},
					{"name": "limit", "type": "Int", "description": "Maximum number of results (default: 20)"},
				},
				"returns": "Array of order objects",
			},
			{
				"name":        "order",
				"description": "Fetch a single order by its ID.",
				"arguments": []map[string]string{
					{"name": "id", "type": "String", "description": "Order ID (required)"},
				},
				"returns": "Order object",
			},
		},
	}

	writeJSON(w, http.StatusOK, schema)
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{
		"errors": []map[string]any{{"message": msg}},
	})
}
