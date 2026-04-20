#!/usr/bin/env bash
# Run tests for all services or a specific domain/service
# Usage: ./run-tests.sh [domain] [service]

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DOMAIN="${1:-}"
SERVICE="${2:-}"
FAILED=()

log() { echo "[$(date '+%H:%M:%S')] $*"; }

run_go_tests() {
  local path="$1" name="$2"
  log "Testing $name (Go)..."
  (cd "$path" && go test ./... -race -count=1 -timeout 60s 2>&1) || FAILED+=("$name")
}

run_node_tests() {
  local path="$1" name="$2"
  log "Testing $name (Node.js)..."
  (cd "$path" && npm test 2>&1) || FAILED+=("$name")
}

run_python_tests() {
  local path="$1" name="$2"
  log "Testing $name (Python)..."
  (cd "$path" && python -m pytest tests/ -q 2>&1) || FAILED+=("$name")
}

run_java_tests() {
  local path="$1" name="$2"
  log "Testing $name (Java/Maven)..."
  (cd "$path" && mvn test -q 2>&1) || FAILED+=("$name")
}

run_kotlin_tests() {
  local path="$1" name="$2"
  log "Testing $name (Kotlin/Gradle)..."
  (cd "$path" && ./gradlew test -q 2>&1) || FAILED+=("$name")
}

test_service() {
  local path="$1"
  local name
  name="$(basename "$path")"

  if [[ -f "$path/go.mod" ]]; then
    run_go_tests "$path" "$name"
  elif [[ -f "$path/package.json" ]]; then
    run_node_tests "$path" "$name"
  elif [[ -f "$path/requirements.txt" ]]; then
    run_python_tests "$path" "$name"
  elif [[ -f "$path/pom.xml" ]]; then
    run_java_tests "$path" "$name"
  elif [[ -f "$path/build.gradle.kts" ]]; then
    run_kotlin_tests "$path" "$name"
  else
    log "SKIP $name — no recognised build file"
  fi
}

if [[ -n "$SERVICE" && -n "$DOMAIN" ]]; then
  test_service "${REPO_ROOT}/src/${DOMAIN}/${SERVICE}"
elif [[ -n "$DOMAIN" ]]; then
  for path in "${REPO_ROOT}/src/${DOMAIN}"/*/; do
    test_service "$path"
  done
else
  for path in "${REPO_ROOT}/src/"**/*/; do
    [[ -f "$path/Dockerfile" ]] && test_service "$path"
  done
fi

if [[ ${#FAILED[@]} -gt 0 ]]; then
  echo ""
  echo "FAILED services:"
  printf '  - %s\n' "${FAILED[@]}"
  exit 1
fi

log "All tests passed."
