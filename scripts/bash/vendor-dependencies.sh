#!/usr/bin/env bash
# vendor-dependencies.sh
# Downloads ALL dependencies into each service's source directory so Docker
# builds never need internet access (base images still need to be pulled once).
#
# Per language, the following directories are created inside each service:
#   Go      → vendor/         (go mod vendor — standard Go vendoring)
#   Python  → wheels/         (pip download — pre-built wheel files)
#   Node.js → node_modules/   (npm ci — full offline node_modules)
#   Rust    → vendor/         (cargo vendor — all crates source)
#   Java    → vendor/m2/      (mvn dep:go-offline with local repo)
#   Kotlin  → vendor/gradle/  (gradlew dep --project-cache-dir)
#   Scala   → vendor/ivy2/    (sbt -ivy target)
#   C#      → vendor/nuget/   (dotnet restore --packages)
#
# After running this script, commit all vendor* directories to git.
# Docker builds will automatically detect and use them (no --build-arg needed).
#
# Usage:
#   ./vendor-dependencies.sh                      # all services
#   ./vendor-dependencies.sh commerce             # one domain
#   ./vendor-dependencies.sh commerce cart-service  # one service

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILTER_DOMAIN="${1:-}"
FILTER_SERVICE="${2:-}"

LOG_DIR="${REPO_ROOT}/logs"
mkdir -p "$LOG_DIR"
RUN_LOG="${LOG_DIR}/vendor-$(date '+%Y%m%d-%H%M%S').log"

TOTAL=0; PASSED=0; FAILED=0
FAILED_LIST=()

log()     { local m="[$(date '+%H:%M:%S')] $*"; echo "$m"; echo "$m" >> "$RUN_LOG"; }
log_raw() { echo "$*" | tee -a "$RUN_LOG"; }
sep()     { log_raw "────────────────────────────────────────────────────────────────"; }

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

sep
log_raw "  ShopOS — Vendor Dependencies (Air-Gap Prep)"
log_raw "  Services : ${TOTAL}"
log_raw "  Log      : ${RUN_LOG}"
sep
echo ""

IDX=0
for SERVICE_PATH in "${VALID_PATHS[@]}"; do
  IDX=$(( IDX + 1 ))
  SERVICE="$(basename "$SERVICE_PATH")"
  LANG=$(detect_lang "$SERVICE_PATH")
  PROGRESS="[${IDX}/${TOTAL}]"

  log "${PROGRESS} ── ${SERVICE} (${LANG})"

  case "$LANG" in

    # ── Go ─────────────────────────────────────────────────────────────────────
    # Creates vendor/ directory — standard Go module vendoring.
    # Dockerfile detects vendor/ and passes -mod=vendor to go build.
    go)
      if [[ -d "${SERVICE_PATH}/vendor" ]]; then
        log "${PROGRESS}   vendor/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      log "${PROGRESS}   go mod tidy && go mod vendor..."
      if (cd "$SERVICE_PATH" \
            && go mod tidy  >> "$RUN_LOG" 2>&1 \
            && go mod vendor >> "$RUN_LOG" 2>&1); then
        COUNT=$(find "${SERVICE_PATH}/vendor" -type d | wc -l)
        log "${PROGRESS} ✓ DONE — ${SERVICE} (${COUNT} vendor packages)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── Python ─────────────────────────────────────────────────────────────────
    # Creates wheels/ directory with .whl files.
    # Dockerfile detects wheels/ and uses --no-index --find-links.
    python)
      if [[ -d "${SERVICE_PATH}/wheels" ]] \
           && [[ -n "$(ls "${SERVICE_PATH}/wheels"/*.whl 2>/dev/null)" ]]; then
        log "${PROGRESS}   wheels/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      mkdir -p "${SERVICE_PATH}/wheels"
      log "${PROGRESS}   pip download..."
      if pip download \
            --dest "${SERVICE_PATH}/wheels" \
            --requirement "${SERVICE_PATH}/requirements.txt" \
            >> "$RUN_LOG" 2>&1; then
        COUNT=$(ls "${SERVICE_PATH}/wheels" | wc -l)
        log "${PROGRESS} ✓ DONE — ${SERVICE} (${COUNT} wheel files)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        rm -rf "${SERVICE_PATH}/wheels"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── Node.js ────────────────────────────────────────────────────────────────
    # Creates node_modules/ using npm ci (respects package-lock.json exactly).
    # Dockerfile detects node_modules/ and copies it in, skipping npm install.
    nodejs)
      if [[ -d "${SERVICE_PATH}/node_modules" ]]; then
        log "${PROGRESS}   node_modules/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      log "${PROGRESS}   npm ci..."
      if (cd "$SERVICE_PATH" && npm ci >> "$RUN_LOG" 2>&1); then
        COUNT=$(ls "${SERVICE_PATH}/node_modules" | wc -l)
        log "${PROGRESS} ✓ DONE — ${SERVICE} (${COUNT} packages in node_modules)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── Rust ───────────────────────────────────────────────────────────────────
    # Creates vendor/ and .cargo/config.toml pointing to it.
    # Dockerfile detects vendor/ and uses CARGO_HOME pointing to it.
    rust)
      if [[ -d "${SERVICE_PATH}/vendor" ]]; then
        log "${PROGRESS}   vendor/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      log "${PROGRESS}   cargo vendor..."
      if (cd "$SERVICE_PATH" && cargo vendor vendor >> "$RUN_LOG" 2>&1); then
        mkdir -p "${SERVICE_PATH}/.cargo"
        cat > "${SERVICE_PATH}/.cargo/config.toml" << 'CARGO_EOF'
[source.crates-io]
replace-with = "vendored-sources"

[source.vendored-sources]
directory = "vendor"
CARGO_EOF
        COUNT=$(ls "${SERVICE_PATH}/vendor" | wc -l)
        log "${PROGRESS} ✓ DONE — ${SERVICE} (${COUNT} crates vendored)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── Java (Maven) ──────────────────────────────────────────────────────────
    # Downloads all Maven artifacts into vendor/m2/ inside the service dir.
    # Dockerfile COPYs vendor/m2 as the local Maven repo and builds offline.
    java)
      if [[ -d "${SERVICE_PATH}/vendor/m2" ]]; then
        log "${PROGRESS}   vendor/m2/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      mkdir -p "${SERVICE_PATH}/vendor/m2"
      log "${PROGRESS}   mvn dependency:go-offline..."
      if (cd "$SERVICE_PATH" \
            && mvn dependency:go-offline \
                 -Dmaven.repo.local="${SERVICE_PATH}/vendor/m2" \
                 -q >> "$RUN_LOG" 2>&1); then
        log "${PROGRESS} ✓ DONE — ${SERVICE} (deps in vendor/m2)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        rm -rf "${SERVICE_PATH}/vendor/m2"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── Kotlin (Gradle) ───────────────────────────────────────────────────────
    # Downloads all Gradle artifacts into vendor/gradle/ inside the service dir.
    # Dockerfile COPYs vendor/gradle as the Gradle project cache.
    kotlin)
      if [[ -d "${SERVICE_PATH}/vendor/gradle" ]]; then
        log "${PROGRESS}   vendor/gradle/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      mkdir -p "${SERVICE_PATH}/vendor/gradle"
      log "${PROGRESS}   gradlew dependencies..."
      if (cd "$SERVICE_PATH" \
            && ./gradlew dependencies \
                 --project-cache-dir "${SERVICE_PATH}/vendor/gradle" \
                 >> "$RUN_LOG" 2>&1); then
        log "${PROGRESS} ✓ DONE — ${SERVICE} (deps in vendor/gradle)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        rm -rf "${SERVICE_PATH}/vendor/gradle"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── Scala (sbt) ───────────────────────────────────────────────────────────
    # Downloads all sbt/ivy2 artifacts into vendor/ivy2/ inside the service dir.
    scala)
      if [[ -d "${SERVICE_PATH}/vendor/ivy2" ]]; then
        log "${PROGRESS}   vendor/ivy2/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      mkdir -p "${SERVICE_PATH}/vendor/ivy2"
      log "${PROGRESS}   sbt update..."
      if (cd "$SERVICE_PATH" \
            && sbt -ivy "${SERVICE_PATH}/vendor/ivy2" update \
               >> "$RUN_LOG" 2>&1); then
        log "${PROGRESS} ✓ DONE — ${SERVICE} (deps in vendor/ivy2)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        rm -rf "${SERVICE_PATH}/vendor/ivy2"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    # ── C# (.NET) ─────────────────────────────────────────────────────────────
    # Downloads all NuGet packages into vendor/nuget/ inside the service dir.
    # Dockerfile COPYs vendor/nuget and restores from there offline.
    dotnet)
      if [[ -d "${SERVICE_PATH}/vendor/nuget" ]]; then
        log "${PROGRESS}   vendor/nuget/ already present — skipping"
        PASSED=$(( PASSED + 1 ))
        continue
      fi
      mkdir -p "${SERVICE_PATH}/vendor/nuget"
      CSPROJ=$(ls "${SERVICE_PATH}"/*.csproj 2>/dev/null | head -1)
      log "${PROGRESS}   dotnet restore..."
      if (cd "$SERVICE_PATH" \
            && dotnet restore "$CSPROJ" \
                 --packages "${SERVICE_PATH}/vendor/nuget" \
                 >> "$RUN_LOG" 2>&1); then
        COUNT=$(ls "${SERVICE_PATH}/vendor/nuget" | wc -l)
        log "${PROGRESS} ✓ DONE — ${SERVICE} (${COUNT} packages in vendor/nuget)"
        PASSED=$(( PASSED + 1 ))
      else
        log "${PROGRESS} ✗ FAILED — ${SERVICE}"
        rm -rf "${SERVICE_PATH}/vendor/nuget"
        FAILED=$(( FAILED + 1 ))
        FAILED_LIST+=("${SERVICE}")
      fi
      ;;

    *)
      log "${PROGRESS} ~ SKIP — ${SERVICE} (unknown language: ${LANG})"
      ;;
  esac
  echo ""
done

sep
log_raw "  Vendor Summary"
log_raw "  Total   : ${TOTAL}"
log_raw "  Passed  : ${PASSED}"
log_raw "  Failed  : ${FAILED}"
sep

if [[ ${#FAILED_LIST[@]} -gt 0 ]]; then
  echo ""
  log_raw "  Failed services:"
  for s in "${FAILED_LIST[@]}"; do
    log_raw "    ✗  ${s}"
  done
fi

echo ""
log "Log saved to: ${RUN_LOG}"
[[ $FAILED -gt 0 ]] && exit 1 || exit 0
