# ============================================================
# ShopOS — Health Check Tests
# Run: hurl --test api-testing/hurl/health-checks.hurl
#      hurl --variable base_host=localhost api-testing/hurl/health-checks.hurl
# ============================================================

# ── api-gateway :8080 ─────────────────────────────────────────────────────────
GET http://localhost:8080/healthz
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.status" == "ok"

# ── web-bff :8081 ─────────────────────────────────────────────────────────────
GET http://localhost:8081/healthz
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.status" == "ok"

# ── mobile-bff :8082 ──────────────────────────────────────────────────────────
GET http://localhost:8082/healthz
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.status" == "ok"

# ── partner-bff :8083 ─────────────────────────────────────────────────────────
GET http://localhost:8083/healthz
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.status" == "ok"

# ── admin-portal :8085 ────────────────────────────────────────────────────────
GET http://localhost:8085/healthz
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"

# ── graphql-gateway :8086 ─────────────────────────────────────────────────────
GET http://localhost:8086/healthz
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"

# ── tenant-service :8087 ──────────────────────────────────────────────────────
GET http://localhost:8087/healthz
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"

# ── supplier-portal-service :8088 ─────────────────────────────────────────────
GET http://localhost:8088/healthz
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"

# ── health-check-service :8090 ────────────────────────────────────────────────
GET http://localhost:8090/healthz
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"

# ── webhook-service :8091 ─────────────────────────────────────────────────────
GET http://localhost:8091/healthz
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"
