package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/enterprise/graphql-gateway/internal/client"
	"github.com/enterprise/graphql-gateway/internal/config"
)

// Resolver dispatches GraphQL field names to downstream HTTP services and merges
// the results into a single response map.
type Resolver struct {
	cfg  *config.Config
	doer client.Doer
}

// New creates a Resolver with the given config and HTTP doer.
func New(cfg *config.Config, doer client.Doer) *Resolver {
	return &Resolver{cfg: cfg, doer: doer}
}

// -------------------------------------------------------------------------
// Individual field resolvers
// -------------------------------------------------------------------------

// Products fetches a paginated list of products from the catalog service.
func (r *Resolver) Products(ctx context.Context, limit, offset int) (any, error) {
	url := fmt.Sprintf("%s/products?limit=%d&offset=%d", r.cfg.CatalogURL, limit, offset)
	return r.getJSON(ctx, url)
}

// Product fetches a single product by ID from the catalog service.
func (r *Resolver) Product(ctx context.Context, id string) (any, error) {
	url := fmt.Sprintf("%s/products/%s", r.cfg.CatalogURL, id)
	return r.getJSON(ctx, url)
}

// Categories fetches all product categories from the catalog service.
func (r *Resolver) Categories(ctx context.Context) (any, error) {
	url := fmt.Sprintf("%s/categories", r.cfg.CatalogURL)
	return r.getJSON(ctx, url)
}

// Cart fetches the active shopping cart for the given user.
func (r *Resolver) Cart(ctx context.Context, userID string) (any, error) {
	url := fmt.Sprintf("%s/carts/%s", r.cfg.CartURL, userID)
	return r.getJSON(ctx, url)
}

// Orders fetches a paginated list of orders for the given user.
func (r *Resolver) Orders(ctx context.Context, userID string, limit int) (any, error) {
	url := fmt.Sprintf("%s/orders?userId=%s&limit=%d", r.cfg.OrdersURL, userID, limit)
	return r.getJSON(ctx, url)
}

// Order fetches a single order by ID.
func (r *Resolver) Order(ctx context.Context, id string) (any, error) {
	url := fmt.Sprintf("%s/orders/%s", r.cfg.OrdersURL, id)
	return r.getJSON(ctx, url)
}

// -------------------------------------------------------------------------
// Query execution
// -------------------------------------------------------------------------

// Execute parses a simplified GraphQL-like query string and dispatches each
// top-level field to the appropriate resolver.  It returns:
//   - data: map of field name → resolved value (nil values are omitted)
//   - errors: list of non-fatal field-level error messages
//
// Supported query format:
//
//	{ products(limit:10,offset:0) categories cart(userId:"u1") orders(userId:"u1",limit:5) product(id:"p1") order(id:"o1") }
func (r *Resolver) Execute(ctx context.Context, query string) (map[string]any, []string) {
	fields := parseQuery(query)

	data := make(map[string]any, len(fields))
	var errors []string

	for _, f := range fields {
		val, err := r.dispatch(ctx, f)
		if err != nil {
			errors = append(errors, fmt.Sprintf("field %q: %s", f.name, err.Error()))
			continue
		}
		data[f.name] = val
	}
	return data, errors
}

// dispatch routes a parsed field to the correct resolver method.
func (r *Resolver) dispatch(ctx context.Context, f field) (any, error) {
	switch f.name {
	case "products":
		limit := argInt(f.args, "limit", 20)
		offset := argInt(f.args, "offset", 0)
		return r.Products(ctx, limit, offset)

	case "product":
		id := argStr(f.args, "id", "")
		if id == "" {
			return nil, fmt.Errorf("argument 'id' is required")
		}
		return r.Product(ctx, id)

	case "categories":
		return r.Categories(ctx)

	case "cart":
		userID := argStr(f.args, "userId", "")
		if userID == "" {
			return nil, fmt.Errorf("argument 'userId' is required")
		}
		return r.Cart(ctx, userID)

	case "orders":
		userID := argStr(f.args, "userId", "")
		if userID == "" {
			return nil, fmt.Errorf("argument 'userId' is required")
		}
		limit := argInt(f.args, "limit", 20)
		return r.Orders(ctx, userID, limit)

	case "order":
		id := argStr(f.args, "id", "")
		if id == "" {
			return nil, fmt.Errorf("argument 'id' is required")
		}
		return r.Order(ctx, id)

	default:
		return nil, fmt.Errorf("unknown field")
	}
}

// -------------------------------------------------------------------------
// HTTP helper
// -------------------------------------------------------------------------

func (r *Resolver) getJSON(ctx context.Context, url string) (any, error) {
	body, status, err := r.doer.Get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("downstream request failed: %w", err)
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("downstream returned status %d", status)
	}
	var result any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON from downstream: %w", err)
	}
	return result, nil
}

// -------------------------------------------------------------------------
// Query parser
// -------------------------------------------------------------------------

// field represents a single top-level selection in the query.
type field struct {
	name string
	args map[string]string
}

// fieldRe matches:  fieldName  or  fieldName(key:val, key:"val", ...)
var fieldRe = regexp.MustCompile(`(\w+)(?:\(([^)]*)\))?`)

// parseQuery extracts the list of top-level fields from a query string like:
//
//	{ products(limit:10) categories cart(userId:"u1") }
func parseQuery(query string) []field {
	// Strip outer braces and whitespace.
	query = strings.TrimSpace(query)
	query = strings.TrimPrefix(query, "{")
	query = strings.TrimSuffix(query, "}")
	query = strings.TrimSpace(query)

	var fields []field
	matches := fieldRe.FindAllStringSubmatch(query, -1)
	for _, m := range matches {
		f := field{
			name: m[1],
			args: parseArgs(m[2]),
		}
		fields = append(fields, f)
	}
	return fields
}

// argRe matches individual key:value pairs inside an argument list.
// Values may be quoted strings or bare tokens.
var argRe = regexp.MustCompile(`(\w+)\s*:\s*(?:"([^"]*)"|([\w.-]+))`)

// parseArgs converts a raw argument string like `limit:10,userId:"u1"` into a
// string→string map.
func parseArgs(raw string) map[string]string {
	result := make(map[string]string)
	if raw == "" {
		return result
	}
	for _, m := range argRe.FindAllStringSubmatch(raw, -1) {
		key := m[1]
		// m[2] is the quoted value, m[3] is the bare value.
		if m[2] != "" {
			result[key] = m[2]
		} else {
			result[key] = m[3]
		}
	}
	return result
}

func argStr(args map[string]string, key, def string) string {
	if v, ok := args[key]; ok {
		return v
	}
	return def
}

func argInt(args map[string]string, key string, def int) int {
	if v, ok := args[key]; ok {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}
