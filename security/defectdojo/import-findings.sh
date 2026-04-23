#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# import-findings.sh — Import CI security scan results into DefectDojo
#
# Usage:
#   export DD_API_KEY="your-defectdojo-api-key"
#   export DD_URL="http://defectdojo:8080"           # or https://defectdojo.shopos.internal
#   export CI_COMMIT_SHA="$(git rev-parse HEAD)"
#   export SERVICE_NAME="api-gateway"                # service being scanned
#   export SCAN_DATE="$(date +%Y-%m-%d)"
#
#   ./import-findings.sh
#
# Artifacts expected (produced by CI steps before this script runs):
#   trivy-results.sarif   — from: trivy image --format sarif
#   grype-results.json    — from: grype ... -o json
#   zap-results.xml       — from: zap-baseline.py -x zap-results.xml
# ──────────────────────────────────────────────────────────────────────────────
set -euo pipefail

# ── Configuration ─────────────────────────────────────────────────────────────
DD_URL="${DD_URL:-http://defectdojo:8080}"
DD_API_KEY="${DD_API_KEY:?DD_API_KEY env var is required}"
SERVICE_NAME="${SERVICE_NAME:-shopos}"
CI_COMMIT_SHA="${CI_COMMIT_SHA:-unknown}"
SCAN_DATE="${SCAN_DATE:-$(date +%Y-%m-%d)}"
BRANCH="${GIT_BRANCH:-main}"
BUILD_ID="${BUILD_NUMBER:-local}"

TRIVY_SARIF="${TRIVY_SARIF:-trivy-results.sarif}"
GRYPE_JSON="${GRYPE_JSON:-grype-results.json}"
ZAP_XML="${ZAP_XML:-zap-results.xml}"

# DefectDojo product and engagement names
PRODUCT_NAME="ShopOS"
ENGAGEMENT_NAME="CI Build ${BUILD_ID} — ${SERVICE_NAME} @ ${CI_COMMIT_SHA:0:8}"

# ── Helper: API call with error handling ──────────────────────────────────────
dd_api() {
  local method="$1"
  local path="$2"
  shift 2
  local response
  response=$(curl --silent --show-error --fail \
    --request "$method" \
    --header "Authorization: Token ${DD_API_KEY}" \
    "${DD_URL}/api/v2${path}" \
    "$@")
  echo "$response"
}

# ── Helper: extract JSON field ────────────────────────────────────────────────
json_field() {
  echo "$1" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d$2)" 2>/dev/null \
    || echo "$1" | grep -o "\"$3\":[^,}]*" | head -1 | sed 's/.*://;s/[^0-9]//g'
}

# ── Step 1: Ensure product exists (create if missing) ─────────────────────────
echo "==> [1/6] Looking up product: ${PRODUCT_NAME}"

PRODUCT_RESPONSE=$(dd_api GET "/products/?name=${PRODUCT_NAME// /%20}&limit=1")
PRODUCT_COUNT=$(echo "$PRODUCT_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['count'])" 2>/dev/null || echo "0")

if [ "$PRODUCT_COUNT" -eq 0 ]; then
  echo "    Product not found — creating..."
  PRODUCT_RESPONSE=$(dd_api POST "/products/" \
    --header "Content-Type: application/json" \
    --data "{
      \"name\": \"${PRODUCT_NAME}\",
      \"description\": \"ShopOS enterprise commerce platform — 224 microservices across 18 domains\",
      \"prod_type\": 1,
      \"business_criticality\": \"very high\",
      \"platform\": \"web service\",
      \"lifecycle\": \"production\",
      \"origin\": \"internal\",
      \"enable_simple_risk_acceptance\": true,
      \"enable_full_risk_acceptance\": true
    }")
  echo "    Product created."
fi

PRODUCT_ID=$(echo "$PRODUCT_RESPONSE" | python3 -c "
import sys, json
d = json.load(sys.stdin)
results = d.get('results', [d])
print(results[0]['id'])
" 2>/dev/null)

echo "    Product ID: ${PRODUCT_ID}"

# ── Step 2: Create engagement ─────────────────────────────────────────────────
echo "==> [2/6] Creating engagement: ${ENGAGEMENT_NAME}"

ENGAGEMENT_RESPONSE=$(dd_api POST "/engagements/" \
  --header "Content-Type: application/json" \
  --data "{
    \"name\": \"${ENGAGEMENT_NAME}\",
    \"product\": ${PRODUCT_ID},
    \"target_start\": \"${SCAN_DATE}\",
    \"target_end\": \"${SCAN_DATE}\",
    \"status\": \"In Progress\",
    \"engagement_type\": \"CI/CD\",
    \"deduplication_on_engagement\": false,
    \"branch_tag\": \"${BRANCH}\",
    \"commit_hash\": \"${CI_COMMIT_SHA}\",
    \"build_id\": \"${BUILD_ID}\",
    \"build_server\": 1,
    \"source_code_management_uri\": \"https://github.com/prabhat-roy/ShopOS\",
    \"description\": \"Automated security scan for service: ${SERVICE_NAME}\"
  }")

ENGAGEMENT_ID=$(echo "$ENGAGEMENT_RESPONSE" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null)
echo "    Engagement ID: ${ENGAGEMENT_ID}"

# ── Step 3: Import Trivy SARIF results ────────────────────────────────────────
echo "==> [3/6] Importing Trivy SARIF results from: ${TRIVY_SARIF}"

if [ -f "${TRIVY_SARIF}" ]; then
  TRIVY_RESPONSE=$(dd_api POST "/import-scan/" \
    --form "scan_type=SARIF" \
    --form "engagement=${ENGAGEMENT_ID}" \
    --form "verified=false" \
    --form "active=true" \
    --form "minimum_severity=Info" \
    --form "close_old_findings=false" \
    --form "push_to_jira=false" \
    --form "file=@${TRIVY_SARIF};type=application/json")
  echo "    Trivy import response: $(echo "$TRIVY_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print('test_id=' + str(d.get('test','N/A')), 'found=' + str(d.get('statistics',{}).get('total_findings', 'N/A')))" 2>/dev/null || echo "done")"
else
  echo "    WARNING: ${TRIVY_SARIF} not found — skipping Trivy import"
fi

# ── Step 4: Import Grype JSON results ─────────────────────────────────────────
echo "==> [4/6] Importing Grype JSON results from: ${GRYPE_JSON}"

if [ -f "${GRYPE_JSON}" ]; then
  GRYPE_RESPONSE=$(dd_api POST "/import-scan/" \
    --form "scan_type=Anchore Grype" \
    --form "engagement=${ENGAGEMENT_ID}" \
    --form "verified=false" \
    --form "active=true" \
    --form "minimum_severity=Info" \
    --form "close_old_findings=false" \
    --form "push_to_jira=false" \
    --form "file=@${GRYPE_JSON};type=application/json")
  echo "    Grype import response: $(echo "$GRYPE_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print('test_id=' + str(d.get('test','N/A')))" 2>/dev/null || echo "done")"
else
  echo "    WARNING: ${GRYPE_JSON} not found — skipping Grype import"
fi

# ── Step 5: Import OWASP ZAP XML results ──────────────────────────────────────
echo "==> [5/6] Importing OWASP ZAP XML results from: ${ZAP_XML}"

if [ -f "${ZAP_XML}" ]; then
  ZAP_RESPONSE=$(dd_api POST "/import-scan/" \
    --form "scan_type=ZAP Scan" \
    --form "engagement=${ENGAGEMENT_ID}" \
    --form "verified=false" \
    --form "active=true" \
    --form "minimum_severity=Info" \
    --form "close_old_findings=false" \
    --form "push_to_jira=false" \
    --form "file=@${ZAP_XML};type=text/xml")
  echo "    ZAP import response: $(echo "$ZAP_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print('test_id=' + str(d.get('test','N/A')))" 2>/dev/null || echo "done")"
else
  echo "    WARNING: ${ZAP_XML} not found — skipping ZAP import"
fi

# ── Step 6: Close engagement ──────────────────────────────────────────────────
echo "==> [6/6] Closing engagement ${ENGAGEMENT_ID}"

dd_api PATCH "/engagements/${ENGAGEMENT_ID}/" \
  --header "Content-Type: application/json" \
  --data "{\"status\": \"Completed\"}" > /dev/null

echo ""
echo "All findings imported successfully."
echo "  Product:    ${PRODUCT_NAME} (ID: ${PRODUCT_ID})"
echo "  Engagement: ${ENGAGEMENT_NAME} (ID: ${ENGAGEMENT_ID})"
echo "  View at:    ${DD_URL}/engagement/${ENGAGEMENT_ID}"
