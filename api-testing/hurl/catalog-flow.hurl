# ============================================================
# ShopOS — Catalog Service Flow Tests
# Covers: list products, get product by ID, search, categories
# Run: hurl --test api-testing/hurl/catalog-flow.hurl
# Prerequisites: auth-flow.hurl must have run (or set access_token variable)
# ============================================================

# ── 1. List products (paginated) ──────────────────────────────────────────────
GET http://localhost:8080/api/v1/products
  ?page=1
  &pageSize=20
  &sort=createdAt:desc
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.data" isCollection
jsonpath "$.pagination.page" == 1
jsonpath "$.pagination.pageSize" == 20
jsonpath "$.pagination.totalItems" isInteger
[Captures]
first_product_id: jsonpath "$.data[0].id"
first_product_sku: jsonpath "$.data[0].sku"

# ── 2. Get product by ID ──────────────────────────────────────────────────────
GET http://localhost:8080/api/v1/products/{{first_product_id}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{first_product_id}}
jsonpath "$.sku" isString
jsonpath "$.name" isString
jsonpath "$.price.amount" isFloat
jsonpath "$.price.currency" isString
jsonpath "$.status" == "active"
jsonpath "$.inventory.available" isInteger

# ── 3. Get product by SKU ─────────────────────────────────────────────────────
GET http://localhost:8080/api/v1/products/sku/{{first_product_sku}}
HTTP 200
[Asserts]
jsonpath "$.sku" == {{first_product_sku}}
jsonpath "$.id" isString

# ── 4. Search products — keyword search ───────────────────────────────────────
GET http://localhost:8080/api/v1/products/search
  ?q=laptop
  &page=1
  &pageSize=10
HTTP 200
[Asserts]
jsonpath "$.data" isCollection
jsonpath "$.query" == "laptop"
jsonpath "$.pagination" isCollection

# ── 5. Search products — with category filter ────────────────────────────────
GET http://localhost:8080/api/v1/products/search
  ?q=laptop
  &category=electronics
  &minPrice=500
  &maxPrice=2000
  &page=1
  &pageSize=10
HTTP 200
[Asserts]
jsonpath "$.data" isCollection
jsonpath "$.filters.category" == "electronics"

# ── 6. List all categories ────────────────────────────────────────────────────
GET http://localhost:8080/api/v1/categories
HTTP 200
[Asserts]
jsonpath "$.data" isCollection
jsonpath "$.data[0].id" isString
jsonpath "$.data[0].name" isString
jsonpath "$.data[0].slug" isString
[Captures]
first_category_id: jsonpath "$.data[0].id"
first_category_slug: jsonpath "$.data[0].slug"

# ── 7. Get category by slug ───────────────────────────────────────────────────
GET http://localhost:8080/api/v1/categories/{{first_category_slug}}
HTTP 200
[Asserts]
jsonpath "$.id" == {{first_category_id}}
jsonpath "$.slug" == {{first_category_slug}}

# ── 8. Get products in category ───────────────────────────────────────────────
GET http://localhost:8080/api/v1/categories/{{first_category_id}}/products
  ?page=1
  &pageSize=10
HTTP 200
[Asserts]
jsonpath "$.data" isCollection

# ── 9. Get product variants ───────────────────────────────────────────────────
GET http://localhost:8080/api/v1/products/{{first_product_id}}/variants
HTTP 200
[Asserts]
jsonpath "$.data" isCollection

# ── 10. Get product inventory ─────────────────────────────────────────────────
GET http://localhost:8080/api/v1/products/{{first_product_id}}/inventory
HTTP 200
[Asserts]
jsonpath "$.productId" == {{first_product_id}}
jsonpath "$.available" isInteger
jsonpath "$.reserved" isInteger
jsonpath "$.total" isInteger

# ── 11. Search with no results (valid response) ───────────────────────────────
GET http://localhost:8080/api/v1/products/search
  ?q=xyzzy-does-not-exist-abc123
HTTP 200
[Asserts]
jsonpath "$.data" isCollection
jsonpath "$.pagination.totalItems" == 0

# ── 12. Request non-existent product → 404 ───────────────────────────────────
GET http://localhost:8080/api/v1/products/00000000-0000-0000-0000-000000000000
HTTP 404
[Asserts]
jsonpath "$.error" isString
