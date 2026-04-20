#!/usr/bin/env bash
# Build all 154 ShopOS Docker images sequentially.
# Images are tagged as shopos/<service-name>:latest
# A build log is written to logs/build-<timestamp>.log
#
# Usage:
#   ./build-all-images.sh              # build everything
#   ./build-all-images.sh commerce     # build one domain only
#   ./build-all-images.sh commerce order-service  # build one service only

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REGISTRY="${REGISTRY:-shopos}"
TAG="${TAG:-latest}"
FILTER_DOMAIN="${1:-}"
FILTER_SERVICE="${2:-}"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="${LOG_DIR}/build-$(date '+%Y%m%d-%H%M%S').log"

TOTAL=0
PASSED=0
FAILED=0
FAILED_LIST=()

# ── helpers ─────────────────────────────────────────────────────────────────

log() {
    local msg="[$(date '+%H:%M:%S')] $*"
    echo "$msg"
    echo "$msg" >> "$LOG_FILE"
}

log_raw() {
    echo "$*" | tee -a "$LOG_FILE"
}

separator() {
    log_raw "────────────────────────────────────────────────────────────────"
}

# ── collect services ─────────────────────────────────────────────────────────

if [[ -n "$FILTER_SERVICE" && -n "$FILTER_DOMAIN" ]]; then
    mapfile -t SERVICE_PATHS < <(find "${REPO_ROOT}/src/${FILTER_DOMAIN}/${FILTER_SERVICE}" -maxdepth 0 -name "Dockerfile" -printf "%h\n" 2>/dev/null || echo "${REPO_ROOT}/src/${FILTER_DOMAIN}/${FILTER_SERVICE}")
elif [[ -n "$FILTER_DOMAIN" ]]; then
    mapfile -t SERVICE_PATHS < <(find "${REPO_ROOT}/src/${FILTER_DOMAIN}" -maxdepth 1 -mindepth 1 -type d | sort)
else
    mapfile -t SERVICE_PATHS < <(find "${REPO_ROOT}/src" -maxdepth 2 -mindepth 2 -type d | sort)
fi

# Filter to only directories that contain a Dockerfile
VALID_PATHS=()
for path in "${SERVICE_PATHS[@]}"; do
    [[ -f "$path/Dockerfile" ]] && VALID_PATHS+=("$path")
done

TOTAL="${#VALID_PATHS[@]}"

# ── header ───────────────────────────────────────────────────────────────────

separator
log_raw "  ShopOS — Docker Image Build"
log_raw "  Registry : ${REGISTRY}"
log_raw "  Tag      : ${TAG}"
log_raw "  Services : ${TOTAL}"
log_raw "  Log      : ${LOG_FILE}"
separator
echo ""

# ── build loop ───────────────────────────────────────────────────────────────

INDEX=0
for path in "${VALID_PATHS[@]}"; do
    INDEX=$(( INDEX + 1 ))
    SERVICE="$(basename "$path")"
    IMAGE="${REGISTRY}/${SERVICE}:${TAG}"
    PROGRESS="[${INDEX}/${TOTAL}]"

    log "${PROGRESS} Building ${SERVICE} ..."

    if docker build \
        --tag "$IMAGE" \
        --file "$path/Dockerfile" \
        "$path" \
        >> "$LOG_FILE" 2>&1; then
        log "${PROGRESS} ✓  ${IMAGE}"
        PASSED=$(( PASSED + 1 ))
    else
        log "${PROGRESS} ✗  FAILED — ${SERVICE}"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("$SERVICE")
    fi
done

# ── summary ──────────────────────────────────────────────────────────────────

echo ""
separator
log_raw "  Build Summary"
log_raw "  Total   : ${TOTAL}"
log_raw "  Success : ${PASSED}"
log_raw "  Failed  : ${FAILED}"
separator

if [[ ${#FAILED_LIST[@]} -gt 0 ]]; then
    echo ""
    log_raw "  Failed services:"
    for s in "${FAILED_LIST[@]}"; do
        log_raw "    ✗  $s"
    done
    echo ""
    log "Full log: ${LOG_FILE}"
    exit 1
fi

echo ""
log "All ${TOTAL} images built successfully."
log "Full log: ${LOG_FILE}"

# ── list images ──────────────────────────────────────────────────────────────

echo ""
echo "Built images:"
docker images "${REGISTRY}/*" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | sort
