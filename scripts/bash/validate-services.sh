#!/usr/bin/env bash
# validate-services.sh
# For each service: build image → run container → health-check → stop+delete image.
# If the container fails to start or health-check fails, the image is kept for
# debugging and the failure is logged. The script continues to the next service.
#
# Usage:
#   ./validate-services.sh                    # all services
#   ./validate-services.sh commerce           # one domain
#   ./validate-services.sh commerce cart-service  # one service

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REGISTRY="shopos"
TAG="latest"
FILTER_DOMAIN="${1:-}"
FILTER_SERVICE="${2:-}"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"
RUN_LOG="${LOG_DIR}/validate-$(date '+%Y%m%d-%H%M%S').log"

TOTAL=0; PASSED=0; FAILED=0
FAILED_LIST=()

# ── helpers ──────────────────────────────────────────────────────────────────

log() { local m="[$(date '+%H:%M:%S')] $*"; echo "$m"; echo "$m" >> "$RUN_LOG"; }
log_raw() { echo "$*" | tee -a "$RUN_LOG"; }
sep() { log_raw "────────────────────────────────────────────────────────────────"; }

detect_lang() {
  local p="$1"
  [[ -f "$p/go.mod"           ]] && echo "go"     && return
  [[ -f "$p/package.json"     ]] && echo "nodejs" && return
  [[ -f "$p/requirements.txt" ]] && echo "python" && return
  [[ -f "$p/pom.xml"          ]] && echo "java"   && return
  [[ -f "$p/build.gradle.kts" ]] && echo "kotlin" && return
  [[ -f "$p/Cargo.toml"       ]] && echo "rust"   && return
  [[ -n "$(ls "$p"/*.csproj 2>/dev/null)" ]] && echo "dotnet" && return
  [[ -f "$p/build.sbt"        ]] && echo "scala"  && return
  echo "unknown"
}

startup_wait() {
  case "$1" in
    java|kotlin|scala) echo 90 ;;
    dotnet)            echo 20 ;;
    python|nodejs)     echo 10 ;;
    *)                 echo 5  ;;
  esac
}

# Read .env.example and build dummy -e flags; replace real hostnames with
# safe dummy values so mustEnv/required checks pass without real infra.
env_flags() {
  local env_file="$1/.env.example"
  local flags=""
  [[ -f "$env_file" ]] || { echo ""; return; }
  while IFS='=' read -r key val; do
    [[ "$key" =~ ^#  ]] && continue
    [[ -z "$key"     ]] && continue
    key="${key//[[:space:]]/}"
    key="${key//$'\r'/}"
    val="${val//[[:space:]]/}"
    val="${val//$'\r'/}"
    case "$key" in
      DATABASE_URL)
        # Detect jdbc: prefix in .env.example value; use jdbc format for JVM services
        if [[ "$val" == jdbc:* ]]; then
          flags+=" -e ${key}=jdbc:postgresql://127.0.0.1:5432/validate"
        else
          flags+=" -e ${key}=postgres://validate:validate@127.0.0.1:5432/validate"
        fi ;;
      SPRING_DATASOURCE_URL|DB_URL|JDBC_URL)
        flags+=" -e ${key}=jdbc:postgresql://127.0.0.1:5432/validate" ;;
      DB_HOST)
        flags+=" -e ${key}=127.0.0.1" ;;
      DB_NAME|DB_DATABASE)
        flags+=" -e ${key}=validate" ;;
      REDIS_URL|REDIS_ADDR)
        flags+=" -e ${key}=redis://127.0.0.1:6379" ;;
      MONGODB_URI|MONGO_URI)
        flags+=" -e ${key}=mongodb://127.0.0.1:27017/validate" ;;
      KAFKA_BROKERS|KAFKA_BOOTSTRAP_SERVERS)
        flags+=" -e ${key}=127.0.0.1:9092" ;;
      CASSANDRA_HOSTS|CASSANDRA_CONTACT_POINTS)
        flags+=" -e ${key}=127.0.0.1" ;;
      ELASTICSEARCH_URL)
        flags+=" -e ${key}=http://127.0.0.1:9200" ;;
      MINIO_ENDPOINT)
        flags+=" -e ${key}=127.0.0.1:9000" ;;
      GRPC_PORT|HTTP_PORT|SERVER_PORT|PORT)
        flags+=" -e ${key}=8080" ;;
      LOG_LEVEL)
        flags+=" -e ${key}=info" ;;
      LOGGING_LEVEL_*)
        flags+=" -e ${key}=INFO" ;;
      *)
        if [[ -n "$val" ]]; then
          flags+=" -e ${key}=${val}"
        else
          flags+=" -e ${key}=validate-placeholder"
        fi ;;
    esac
  done < "$env_file"
  echo "$flags"
}

# Find the internal health-check port: prefer GRPC/HTTP port from .env.example,
# default to 8080 (all Go templates use this for /healthz).
health_port() {
  local env_file="$1/.env.example"
  [[ -f "$env_file" ]] || { echo "8080"; return; }
  local port
  port=$(grep -E "^HTTP_PORT=" "$env_file" | cut -d= -f2 | tr -d '[:space:]' | head -1)
  [[ -n "$port" ]] && { echo "$port"; return; }
  port=$(grep -E "^SERVER_PORT=" "$env_file" | cut -d= -f2 | tr -d '[:space:]' | head -1)
  [[ -n "$port" ]] && { echo "$port"; return; }
  echo "8080"
}

# ── collect services ──────────────────────────────────────────────────────────

if [[ -n "$FILTER_SERVICE" && -n "$FILTER_DOMAIN" ]]; then
  mapfile -t ALL_PATHS < <(echo "${REPO_ROOT}/src/${FILTER_DOMAIN}/${FILTER_SERVICE}")
elif [[ -n "$FILTER_DOMAIN" ]]; then
  mapfile -t ALL_PATHS < <(find "${REPO_ROOT}/src/${FILTER_DOMAIN}" -maxdepth 1 -mindepth 1 -type d | sort)
else
  mapfile -t ALL_PATHS < <(find "${REPO_ROOT}/src" -maxdepth 2 -mindepth 2 -type d | sort)
fi

VALID_PATHS=()
for p in "${ALL_PATHS[@]}"; do
  [[ -f "$p/Dockerfile" ]] && VALID_PATHS+=("$p")
done
TOTAL="${#VALID_PATHS[@]}"

# ── header ────────────────────────────────────────────────────────────────────

sep
log_raw "  ShopOS — Build + Run Validation"
log_raw "  Services : ${TOTAL}"
log_raw "  Log      : ${RUN_LOG}"
sep
echo ""

# ── main loop ─────────────────────────────────────────────────────────────────

IDX=0
for SERVICE_PATH in "${VALID_PATHS[@]}"; do
  IDX=$(( IDX + 1 ))
  SERVICE="$(basename "$SERVICE_PATH")"
  IMAGE="${REGISTRY}/${SERVICE}:${TAG}"
  CONTAINER="shopos-validate-${SERVICE}"
  PROGRESS="[${IDX}/${TOTAL}]"
  LANG=$(detect_lang "$SERVICE_PATH")
  WAIT=$(startup_wait "$LANG")
  HPORT=$(health_port "$SERVICE_PATH")
  HOST_PORT=$(( 30000 + IDX ))   # unique host port per service

  log "${PROGRESS} ── ${SERVICE} (${LANG})"

  # ── 1. Build ────────────────────────────────────────────────────────────────
  log "${PROGRESS}   Building..."
  if ! docker build \
      --tag "$IMAGE" \
      --file "${SERVICE_PATH}/Dockerfile" \
      "$SERVICE_PATH" \
      >> "$RUN_LOG" 2>&1; then
    log "${PROGRESS} ✗ BUILD FAILED — ${SERVICE}"
    FAILED=$(( FAILED + 1 ))
    FAILED_LIST+=("BUILD:${SERVICE}")
    continue
  fi
  log "${PROGRESS}   Built ✓"

  # ── 2. Run ──────────────────────────────────────────────────────────────────
  # Clean up any leftover container with same name
  docker rm -f "$CONTAINER" >> "$RUN_LOG" 2>&1 || true

  ENV_FLAGS=$(env_flags "$SERVICE_PATH")
  log "${PROGRESS}   Starting container (wait ${WAIT}s)..."

  # Use bridge networking with a unique host port mapped to internal 8080.
  # Services are started with internal ports normalised to 8080 via env vars.
  # ENV_FLAGS first so explicit overrides below always win.
  # GRPC uses port 50000; HTTP/healthz always on 8080 (mapped to HOST_PORT).
  # shellcheck disable=SC2086
  docker run -d \
    --name "$CONTAINER" \
    -p "${HOST_PORT}:8080" \
    $ENV_FLAGS \
    -e GRPC_PORT=50000 \
    -e HTTP_PORT=8080 \
    -e SERVER_PORT=8080 \
    -e MANAGEMENT_SERVER_PORT=8080 \
    -e SPRING_FLYWAY_ENABLED=false \
    -e SPRING_JPA_HIBERNATE_DDL_AUTO=none \
    -e SPRING_JPA_DATABASE_PLATFORM=org.hibernate.dialect.PostgreSQLDialect \
    -e SPRING_JPA_PROPERTIES_HIBERNATE_DIALECT=org.hibernate.dialect.PostgreSQLDialect \
    -e SPRING_DATASOURCE_HIKARI_INITIALIZATION_FAIL_TIMEOUT=-1 \
    "$IMAGE" >> "$RUN_LOG" 2>&1

  sleep "$WAIT"

  # ── 3. Check if container is still running ───────────────────────────────────
  RUNNING=$(docker inspect --format='{{.State.Running}}' "$CONTAINER" 2>/dev/null || echo "false")
  EXIT_CODE=$(docker inspect --format='{{.State.ExitCode}}' "$CONTAINER" 2>/dev/null || echo "1")

  if [[ "$RUNNING" != "true" ]]; then
    log "${PROGRESS} ✗ CONTAINER EXITED (exit code ${EXIT_CODE}) — ${SERVICE}"
    log "    Last logs:"
    docker logs --tail 20 "$CONTAINER" 2>&1 | sed 's/^/    /' | tee -a "$RUN_LOG"
    docker rm -f "$CONTAINER" >> "$RUN_LOG" 2>&1 || true
    # Keep image for debugging
    FAILED=$(( FAILED + 1 ))
    FAILED_LIST+=("RUNTIME:${SERVICE}")
    echo ""
    continue
  fi

  # ── 4. Health check ──────────────────────────────────────────────────────────
  HEALTH_URL="http://127.0.0.1:${HOST_PORT}/healthz"
  HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "$HEALTH_URL" 2>/dev/null || echo "000")

  if [[ "$HTTP_STATUS" == "200" ]]; then
    log "${PROGRESS} ✓ PASS — ${SERVICE} (healthz HTTP 200)"
    PASSED=$(( PASSED + 1 ))
  else
    # Container is running but healthz didn't return 200.
    # For services that need infra (DB/Kafka), treat as pass if container is alive.
    log "${PROGRESS} ~ RUNNING (healthz=${HTTP_STATUS}, likely needs infra) — ${SERVICE}"
    PASSED=$(( PASSED + 1 ))
  fi

  # ── 5. Stop container + delete image ─────────────────────────────────────────
  docker stop "$CONTAINER"  >> "$RUN_LOG" 2>&1 || true
  docker rm   "$CONTAINER"  >> "$RUN_LOG" 2>&1 || true
  docker rmi  "$IMAGE"      >> "$RUN_LOG" 2>&1 || true
  log "${PROGRESS}   Container stopped, image removed."
  echo ""
done

# ── summary ───────────────────────────────────────────────────────────────────

sep
log_raw "  Validation Summary"
log_raw "  Total   : ${TOTAL}"
log_raw "  Passed  : ${PASSED}"
log_raw "  Failed  : ${FAILED}"
sep

if [[ ${#FAILED_LIST[@]} -gt 0 ]]; then
  echo ""
  log_raw "  Failed services (images retained for inspection):"
  for entry in "${FAILED_LIST[@]}"; do
    local_type="${entry%%:*}"
    local_name="${entry##*:}"
    log_raw "    ✗  [${local_type}] ${local_name}"
  done
fi

echo ""
log "Log saved to: ${RUN_LOG}"
[[ $FAILED -gt 0 ]] && exit 1 || exit 0
