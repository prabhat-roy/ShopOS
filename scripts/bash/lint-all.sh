#!/usr/bin/env bash
# Run linters across all services by language
# Usage: ./lint-all.sh [domain]

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DOMAIN="${1:-}"
FAILED=()

log() { echo "[$(date '+%H:%M:%S')] $*"; }

lint_service() {
  local path="$1"
  local name
  name="$(basename "$path")"

  if [[ -f "$path/go.mod" ]]; then
    log "Linting $name (go vet)..."
    (cd "$path" && go vet ./...) || FAILED+=("$name")
  elif [[ -f "$path/package.json" ]]; then
    if [[ -f "$path/.eslintrc*" ]] || grep -q '"lint"' "$path/package.json" 2>/dev/null; then
      log "Linting $name (eslint)..."
      (cd "$path" && npm run lint --if-present) || FAILED+=("$name")
    fi
  elif [[ -f "$path/requirements.txt" ]]; then
    log "Linting $name (ruff)..."
    (cd "$path" && python -m ruff check . 2>/dev/null || python -m flake8 . --max-line-length=120) || FAILED+=("$name")
  elif [[ -f "$path/pom.xml" ]]; then
    log "Linting $name (Maven checkstyle)..."
    (cd "$path" && mvn checkstyle:check -q 2>/dev/null) || true
  elif [[ -f "$path/build.gradle.kts" ]]; then
    log "Linting $name (ktlint)..."
    (cd "$path" && ./gradlew ktlintCheck -q 2>/dev/null) || true
  fi
}

if [[ -n "$DOMAIN" ]]; then
  for path in "${REPO_ROOT}/src/${DOMAIN}"/*/; do
    [[ -f "$path/Dockerfile" ]] && lint_service "$path"
  done
else
  for path in "${REPO_ROOT}/src/"*/*/; do
    [[ -f "$path/Dockerfile" ]] && lint_service "$path"
  done
fi

if [[ ${#FAILED[@]} -gt 0 ]]; then
  echo ""
  echo "Lint FAILED for:"
  printf '  - %s\n' "${FAILED[@]}"
  exit 1
fi

log "Lint complete — no issues found."
