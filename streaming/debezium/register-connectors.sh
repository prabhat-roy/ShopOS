#!/bin/bash
# Register all Debezium connectors via Kafka Connect REST API
set -euo pipefail

CONNECT_URL="${KAFKA_CONNECT_URL:-http://localhost:8083}"

register() {
  local name=$1
  local file=$2
  echo "Registering connector: $name"
  curl -sf -X POST "$CONNECT_URL/connectors" \
    -H "Content-Type: application/json" \
    -d @"$file"
  echo ""
}

update() {
  local name=$1
  local file=$2
  echo "Updating connector: $name"
  curl -sf -X PUT "$CONNECT_URL/connectors/$name/config" \
    -H "Content-Type: application/json" \
    -d "$(jq '.config' "$file")"
  echo ""
}

upsert() {
  local name=$1
  local file=$2
  if curl -sf "$CONNECT_URL/connectors/$name" > /dev/null 2>&1; then
    update "$name" "$file"
  else
    register "$name" "$file"
  fi
}

DIR="$(cd "$(dirname "$0")" && pwd)"

upsert "postgres-orders-connector"  "$DIR/postgres-orders-connector.json"
upsert "mongodb-catalog-connector"  "$DIR/mongodb-catalog-connector.json"

echo "All connectors registered."
echo "Status:"
curl -sf "$CONNECT_URL/connectors?expand=status" | jq '[to_entries[] | {name: .key, state: .value.status.connector.state}]'
