#!/usr/bin/env bash
# parallel-validate.sh — build + run + healthz all services in parallel
# Usage: ./parallel-validate.sh [WORKERS]   (default: 6)

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
REGISTRY="shopos"
TAG="latest"
WORKERS="${1:-6}"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"
RUN_LOG="${LOG_DIR}/parallel-validate-$(date '+%Y%m%d-%H%M%S').log"
TMPDIR_SHARED="$(mktemp -d)"
echo 0 > "${TMPDIR_SHARED}/passed"
echo 0 > "${TMPDIR_SHARED}/failed"
> "${TMPDIR_SHARED}/failed_list"

log_main() { local m="[$(date '+%H:%M:%S')] $*"; echo "$m" | tee -a "$RUN_LOG"; }

# Collect all services
mapfile -t ALL_PATHS < <(find "${REPO_ROOT}/src" -maxdepth 2 -mindepth 2 -type d | sort)
VALID_PATHS=()
for p in "${ALL_PATHS[@]}"; do
  [[ -f "$p/Dockerfile" ]] && VALID_PATHS+=("$p")
done
TOTAL="${#VALID_PATHS[@]}"

log_main "ShopOS Parallel Validate — ${TOTAL} services, ${WORKERS} workers"
log_main "Log: ${RUN_LOG}"

# Write worker script
WORKER_SCRIPT="${TMPDIR_SHARED}/worker.sh"
cat > "$WORKER_SCRIPT" << 'WORKER_EOF'
#!/usr/bin/env bash
set -uo pipefail

SERVICE_PATH="$1"
IDX="$2"
TOTAL="$3"
REGISTRY="$4"
TAG="$5"
LOG_DIR="$6"
RUN_LOG="$7"
TMPDIR_SHARED="$8"

SERVICE="$(basename "$SERVICE_PATH")"
HOST_PORT=$(( 31000 + IDX ))
IMAGE="${REGISTRY}/${SERVICE}:${TAG}"
CONTAINER="shopos-pval-${SERVICE}"
svclog="${LOG_DIR}/svc-${SERVICE}.log"

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
    java|kotlin|scala) echo 120 ;;
    dotnet)            echo 20 ;;
    python|nodejs)     echo 10 ;;
    *)                 echo 5  ;;
  esac
}

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
        if [[ "$val" == jdbc:* ]]; then
          flags+=" -e ${key}=jdbc:postgresql://127.0.0.1:5432/validate"
        else
          flags+=" -e ${key}=postgres://validate:validate@127.0.0.1:5432/validate"
        fi ;;
      SPRING_DATASOURCE_URL|DB_URL|JDBC_URL)
        flags+=" -e ${key}=jdbc:postgresql://127.0.0.1:5432/validate" ;;
      DB_HOST)        flags+=" -e ${key}=127.0.0.1" ;;
      DB_NAME|DB_DATABASE) flags+=" -e ${key}=validate" ;;
      REDIS_URL|REDIS_ADDR) flags+=" -e ${key}=redis://127.0.0.1:6379" ;;
      MONGODB_URI|MONGO_URI) flags+=" -e ${key}=mongodb://127.0.0.1:27017/validate" ;;
      KAFKA_BROKERS|KAFKA_BOOTSTRAP_SERVERS) flags+=" -e ${key}=127.0.0.1:9092" ;;
      CASSANDRA_HOSTS|CASSANDRA_CONTACT_POINTS) flags+=" -e ${key}=127.0.0.1" ;;
      ELASTICSEARCH_URL) flags+=" -e ${key}=http://127.0.0.1:9200" ;;
      MINIO_ENDPOINT) flags+=" -e ${key}=127.0.0.1:9000" ;;
      GRPC_PORT|HTTP_PORT|SERVER_PORT|PORT) flags+=" -e ${key}=8080" ;;
      LOG_LEVEL)      flags+=" -e ${key}=info" ;;
      LOGGING_LEVEL_*) flags+=" -e ${key}=INFO" ;;
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

LANG=$(detect_lang "$SERVICE_PATH")
WAIT=$(startup_wait "$LANG")

echo "[$(date '+%H:%M:%S')] [${IDX}/${TOTAL}] START ${SERVICE} (${LANG})" | tee -a "$RUN_LOG"

# Build
if ! docker build --tag "$IMAGE" --file "${SERVICE_PATH}/Dockerfile" "$SERVICE_PATH" > "$svclog" 2>&1; then
  echo "[$(date '+%H:%M:%S')] [${IDX}/${TOTAL}] ✗ BUILD FAILED — ${SERVICE}" | tee -a "$RUN_LOG"
  tail -5 "$svclog" | sed "s/^/  /" | tee -a "$RUN_LOG"
  echo "BUILD:${SERVICE}" >> "${TMPDIR_SHARED}/failed_list"
  ( flock 9; v=$(cat "${TMPDIR_SHARED}/failed"); echo $(( v + 1 )) > "${TMPDIR_SHARED}/failed" ) 9>"${TMPDIR_SHARED}/failed.lock"
  exit 0
fi

docker rm -f "$CONTAINER" >> "$svclog" 2>&1 || true

ENV_FLAGS=$(env_flags "$SERVICE_PATH")

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
  "$IMAGE" >> "$svclog" 2>&1

sleep "$WAIT"

RUNNING=$(docker inspect --format='{{.State.Running}}' "$CONTAINER" 2>/dev/null || echo "false")
EXIT_CODE=$(docker inspect --format='{{.State.ExitCode}}' "$CONTAINER" 2>/dev/null || echo "1")

if [[ "$RUNNING" != "true" ]]; then
  echo "[$(date '+%H:%M:%S')] [${IDX}/${TOTAL}] ✗ CONTAINER EXITED (exit code ${EXIT_CODE}) — ${SERVICE}" | tee -a "$RUN_LOG"
  docker logs --tail 10 "$CONTAINER" 2>&1 | sed "s/^/  [$SERVICE] /" | tee -a "$RUN_LOG"
  docker rm -f "$CONTAINER" >> "$svclog" 2>&1 || true
  docker rmi "$IMAGE" >> "$svclog" 2>&1 || true
  echo "RUNTIME:${SERVICE}" >> "${TMPDIR_SHARED}/failed_list"
  ( flock 9; v=$(cat "${TMPDIR_SHARED}/failed"); echo $(( v + 1 )) > "${TMPDIR_SHARED}/failed" ) 9>"${TMPDIR_SHARED}/failed.lock"
  exit 0
fi

HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 "http://localhost:${HOST_PORT}/healthz" 2>/dev/null || echo "000")

if [[ "$HTTP_CODE" == "200" ]]; then
  echo "[$(date '+%H:%M:%S')] [${IDX}/${TOTAL}] ✓ PASS — ${SERVICE} (healthz HTTP 200)" | tee -a "$RUN_LOG"
else
  echo "[$(date '+%H:%M:%S')] [${IDX}/${TOTAL}] ~ RUNNING — ${SERVICE} (healthz HTTP ${HTTP_CODE})" | tee -a "$RUN_LOG"
fi

docker rm -f "$CONTAINER" >> "$svclog" 2>&1 || true
docker rmi "$IMAGE" >> "$svclog" 2>&1 || true
( flock 9; v=$(cat "${TMPDIR_SHARED}/passed"); echo $(( v + 1 )) > "${TMPDIR_SHARED}/passed" ) 9>"${TMPDIR_SHARED}/passed.lock"
WORKER_EOF
chmod +x "$WORKER_SCRIPT"

# Build args file
ARGS_FILE="${TMPDIR_SHARED}/args"
IDX=0
for p in "${VALID_PATHS[@]}"; do
  IDX=$(( IDX + 1 ))
  printf '%s\t%d\t%d\n' "$p" "$IDX" "$TOTAL"
done > "$ARGS_FILE"

# Run in parallel using xargs
export WORKER_SCRIPT REGISTRY TAG LOG_DIR RUN_LOG TMPDIR_SHARED TOTAL
cat "$ARGS_FILE" | xargs -P "$WORKERS" -I '{}' bash -c '
  IFS=$'"'"'\t'"'"' read -r spath idx total <<< "{}"
  bash "$WORKER_SCRIPT" "$spath" "$idx" "$total" "$REGISTRY" "$TAG" "$LOG_DIR" "$RUN_LOG" "$TMPDIR_SHARED"
'

PASSED=$(cat "${TMPDIR_SHARED}/passed")
FAILED=$(cat "${TMPDIR_SHARED}/failed")
log_main "════════════════════════════════════"
log_main "DONE — ${PASSED} passed  ${FAILED} failed  (total: $(( PASSED + FAILED ))/${TOTAL})"
if [[ -s "${TMPDIR_SHARED}/failed_list" ]]; then
  log_main "Failed:"
  sort "${TMPDIR_SHARED}/failed_list" | tee -a "$RUN_LOG"
fi
rm -rf "$TMPDIR_SHARED"
