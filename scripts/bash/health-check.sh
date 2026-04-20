#!/usr/bin/env bash
# Check /healthz endpoints for all running services
# Usage: ./health-check.sh [base_url]
# Example: ./health-check.sh http://localhost

set -euo pipefail

BASE="${1:-http://localhost}"

# Map of service -> HTTP port
declare -A SERVICES=(
  [api-gateway]=8080
  [web-bff]=8081
  [mobile-bff]=8082
  [partner-bff]=8083
  [health-check-service]=8090
  [webhook-service]=8091
  [load-generator]=8089
  [admin-portal]=8085
  [graphql-gateway]=8086
  [supplier-portal-service]=8088
  [chatbot-service]=8193
  [attribution-service]=8194
  [clv-service]=8195
  [search-analytics-service]=8196
)

PASS=0
FAIL=0
DOWN=()

log() { echo "[$(date '+%H:%M:%S')] $*"; }

for name in "${!SERVICES[@]}"; do
  port="${SERVICES[$name]}"
  url="${BASE}:${port}/healthz"
  status=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "$url" 2>/dev/null || echo "000")
  if [[ "$status" == "200" ]]; then
    log "OK    $name ($url)"
    ((PASS++))
  else
    log "FAIL  $name ($url) — HTTP $status"
    DOWN+=("$name")
    ((FAIL++))
  fi
done

echo ""
echo "Results: ${PASS} healthy, ${FAIL} unhealthy"

if [[ ${#DOWN[@]} -gt 0 ]]; then
  echo "Unhealthy services:"
  printf '  - %s\n' "${DOWN[@]}"
  exit 1
fi
