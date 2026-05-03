#!/usr/bin/env bash
# Purge CDN + Varnish + Cloudflare cache for a given service or path.
# Used by deployment runbook step "Frontend rollback".
set -euo pipefail
TARGET="${1:?storefront|admin|seller|partner|developer|/path/to/asset}"

# Varnish purge (in-cluster)
kubectl exec -n networking sts/varnish -- varnishadm "ban req.url ~ ^/${TARGET}.*" || true

# Cloudflare API purge
if [[ -n "${CF_API_TOKEN:-}" ]]; then
  curl -sS -X POST "https://api.cloudflare.com/client/v4/zones/${CF_ZONE_ID}/purge_cache" \
    -H "Authorization: Bearer ${CF_API_TOKEN}" \
    -H "Content-Type: application/json" \
    --data "{\"prefixes\":[\"https://shop.example.com/${TARGET}\"]}"
fi

# Fastly purge
if [[ -n "${FASTLY_API_TOKEN:-}" ]]; then
  curl -sS -X POST "https://api.fastly.com/service/${FASTLY_SERVICE_ID}/purge/${TARGET}" \
    -H "Fastly-Key: ${FASTLY_API_TOKEN}"
fi
