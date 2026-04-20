#!/usr/bin/env bash
# Push all built images to a container registry
# Usage: ./push-all.sh [domain]
# Requires REGISTRY env var (default: localhost:5000/shopos)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REGISTRY="${REGISTRY:-localhost:5000/shopos}"
TAG="${TAG:-latest}"
DOMAIN="${1:-}"

log() { echo "[$(date '+%H:%M:%S')] $*"; }

if [[ -n "$DOMAIN" ]]; then
  SEARCH_ROOT="${REPO_ROOT}/src/${DOMAIN}"
else
  SEARCH_ROOT="${REPO_ROOT}/src"
fi

mapfile -t SERVICES < <(find "$SEARCH_ROOT" -maxdepth 2 -name "Dockerfile" -printf "%h\n" | sort)

log "Pushing ${#SERVICES[@]} images to ${REGISTRY}"

for path in "${SERVICES[@]}"; do
  name="$(basename "$path")"
  local_tag="shopos/${name}:${TAG}"
  remote_tag="${REGISTRY}/${name}:${TAG}"
  docker tag "$local_tag" "$remote_tag"
  docker push "$remote_tag"
  log "Pushed $remote_tag"
done

log "All images pushed."
