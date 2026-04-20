#!/usr/bin/env bash
# Run database migrations for all Postgres-backed services
# Usage: ./db-migrate.sh [up|down] [domain] [service]
# Requires: golang-migrate (migrate CLI)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DIRECTION="${1:-up}"
DOMAIN="${2:-}"
SERVICE="${3:-}"
DB_HOST="${DB_HOST:-localhost}"
DB_USER="${DB_USER:-postgres}"
DB_PASS="${DB_PASS:-postgres}"
DB_PORT="${DB_PORT:-5432}"

log() { echo "[$(date '+%H:%M:%S')] $*"; }

if ! command -v migrate &>/dev/null; then
  echo "ERROR: golang-migrate not found. Install: https://github.com/golang-migrate/migrate"
  exit 1
fi

migrate_service() {
  local path="$1"
  local name
  name="$(basename "$path")"
  local migrations_dir="${path}/migrations"

  [[ -d "$migrations_dir" ]] || return 0

  local db_name
  db_name="${name//-/_}"
  local db_url="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${db_name}?sslmode=disable"

  log "Migrating $name ($DIRECTION)..."
  migrate -path "$migrations_dir" -database "$db_url" "$DIRECTION" || {
    log "WARNING: migration failed for $name"
  }
}

if [[ -n "$SERVICE" && -n "$DOMAIN" ]]; then
  migrate_service "${REPO_ROOT}/src/${DOMAIN}/${SERVICE}"
elif [[ -n "$DOMAIN" ]]; then
  for path in "${REPO_ROOT}/src/${DOMAIN}"/*/; do
    migrate_service "$path"
  done
else
  for path in "${REPO_ROOT}/src/"*/*/; do
    [[ -f "$path/Dockerfile" ]] && migrate_service "$path"
  done
fi

log "Migration run complete (direction: $DIRECTION)."
