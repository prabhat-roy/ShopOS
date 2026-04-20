#!/usr/bin/env bash
# Build Docker images for all services (or a specific domain)
# Usage: ./build-all.sh [domain]
# Example: ./build-all.sh commerce

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REGISTRY="${REGISTRY:-shopos}"
TAG="${TAG:-latest}"
DOMAIN="${1:-}"
PARALLEL="${PARALLEL:-4}"

log() { echo "[$(date '+%H:%M:%S')] $*"; }

build_service() {
  local path="$1"
  local name
  name="$(basename "$path")"
  log "Building $name..."
  docker build -t "${REGISTRY}/${name}:${TAG}" "$path" --quiet
  log "Done: ${REGISTRY}/${name}:${TAG}"
}

export -f build_service log
export REGISTRY TAG

if [[ -n "$DOMAIN" ]]; then
  SEARCH_ROOT="${REPO_ROOT}/src/${DOMAIN}"
  if [[ ! -d "$SEARCH_ROOT" ]]; then
    echo "ERROR: Domain '$DOMAIN' not found under src/"
    exit 1
  fi
else
  SEARCH_ROOT="${REPO_ROOT}/src"
fi

mapfile -t SERVICES < <(find "$SEARCH_ROOT" -maxdepth 2 -name "Dockerfile" -printf "%h\n" | sort)

log "Found ${#SERVICES[@]} services to build (parallelism=${PARALLEL})"
printf '%s\n' "${SERVICES[@]}" | xargs -P "$PARALLEL" -I{} bash -c 'build_service "$@"' _ {}

log "All builds complete."
