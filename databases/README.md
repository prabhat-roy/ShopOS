# Databases — ShopOS

Schema definitions and Helm-deployed managed instances for every database used by ShopOS.
Primary OLTP storage (Postgres, MongoDB, Redis) is per-service via Helm + Vault dynamic
credentials; this directory holds the cross-service analytical, vector, graph, and search stores.

---

## Layout

```
databases/
├── postgres/             Flyway migrations — V001-V011 covering 11 domain schemas
│                          (identity, commerce, financial, catalog, supply_chain, b2b,
│                           marketplace, platform, cx, compliance, sustainability, events,
│                           auction, rental, gamification)
├── mongodb/              Document validators + indexes (product catalog, reviews, CMS)
├── redis/                Key patterns + TTL reference per domain
├── clickhouse/           OLAP — orders, events, revenue_daily MV, product_clicks
├── weaviate/             Vector — Product, UserQuery classes (RAG, semantic search)
├── neo4j/                Graph — product recommendation graph
├── dgraph/               Distributed graph DB (Neo4j alternative for horizontal scale)
├── scylladb/             High-throughput Cassandra-compatible time-series
├── opensearch/           Log/audit/security index templates + ILM policies
├── timescaledb/          Time-series Postgres extension — service metrics, inventory, page views
├── memcached/            High-throughput simple cache
├── lakefs/               Git-like data versioning over MinIO (analytics + dbt reproducibility)
└── yugabytedb/           Distributed Postgres-compatible SQL (geo-distributed alternative to CockroachDB)
```

---

## Postgres Flyway migrations

| File | Schema(s) | Notes |
|---|---|---|
| [V001__identity_schema.sql](postgres/V001__identity_schema.sql) | `identity` | users, sessions, MFA, API keys |
| [V002__commerce_schema.sql](postgres/V002__commerce_schema.sql) | `commerce` | orders, payments, cart, fraud |
| [V003__financial_schema.sql](postgres/V003__financial_schema.sql) | `financial` | invoices, accounting, payouts |
| [V004__catalog_schema.sql](postgres/V004__catalog_schema.sql) | `catalog` | category (ltree), brand, price, inventory, bundle, variant, label |
| [V005__supply_chain_schema.sql](postgres/V005__supply_chain_schema.sql) | `supply_chain` | vendor, warehouse, PO, fulfillment, customs, route_plan |
| [V006__b2b_schema.sql](postgres/V006__b2b_schema.sql) | `b2b` | organization, contract, quote, RFP, approval, credit |
| [V007__marketplace_schema.sql](postgres/V007__marketplace_schema.sql) | `marketplace` | seller, listing, commission, dispute, payout |
| [V008__platform_schema.sql](postgres/V008__platform_schema.sql) | `platform` | saga, event_store, webhook, scheduler, feature_flag, tenant |
| [V009__customer_experience_schema.sql](postgres/V009__customer_experience_schema.sql) | `cx` | wishlist, support_ticket, survey, gift_registry, feedback |
| [V010__compliance_sustainability_schema.sql](postgres/V010__compliance_sustainability_schema.sql) | `compliance`, `sustainability` | retention, consent_audit, privacy_request, carbon, eco_score |
| [V011__events_auction_rental_gamification_schema.sql](postgres/V011__events_auction_rental_gamification_schema.sql) | `events_ticketing`, `auction`, `rental`, `gamification` | venue, ticket, auction, bid, lease, badge, challenge |

Apply via Flyway in CI before any service deploy:

```bash
docker run --rm -v $PWD/databases/postgres:/flyway/sql flyway/flyway:10 \
  -url=jdbc:postgresql://postgres-primary.databases.svc:5432/postgres \
  -user=$FLYWAY_USER -password=$FLYWAY_PASSWORD migrate
```

Dynamic credentials per service domain are issued by Vault — see
[`../security/vault/bootstrap/02-secret-engines.sh`](../security/vault/bootstrap/02-secret-engines.sh).

---

## ClickHouse

OLAP analytics — order aggregations, revenue reporting, funnel analysis. Replicated MergeTree
tables with materialized views for daily revenue.

```bash
clickhouse-client --host clickhouse.databases.svc --port 9000 --user shopos
```

## Weaviate / Dgraph

| Tool | Use | Why two? |
|---|---|---|
| Weaviate | Vector search (RAG, semantic product search) | Strong at vector ops + multi-modal |
| Dgraph | Distributed graph at scale | Horizontal scaling beyond Neo4j single-instance |
| Neo4j | Recommendation graph (canonical) | Best Cypher tooling for product team |

## ScyllaDB

High-throughput Cassandra-compatible store for time-series events, session data, IoT
telemetry. Keyspace `shopos_analytics`, RF=3, NetworkTopologyStrategy.

## OpenSearch

Log + audit + security event search. Index templates with ILM policies (logs 30d, audit
365d, security 90d). Used by SIEM (Wazuh) and engineer log debugging.

## TimescaleDB

Time-series Postgres extension for service metrics, inventory events, page views. Hypertables
auto-partition by time.

## LakeFS

Git-like data versioning over MinIO (S3-compatible). Used by dbt + analytics team for
reproducible data runs (branch the data lake, run experiment, merge or discard).

## YugabyteDB

Distributed Postgres-compatible SQL. Used when geo-distributed ACID is required and
CockroachDB licensing is undesired.

---

## Applying schemas (full bootstrap)

```bash
# Postgres
docker run --rm -v $PWD/databases/postgres:/flyway/sql flyway/flyway:10 migrate

# MongoDB
mongosh "$MONGO_URI" databases/mongodb/product-catalog-schema.js

# ClickHouse
clickhouse-client --host clickhouse.databases.svc < databases/clickhouse/schema.sql

# Weaviate
curl -X POST http://weaviate.databases.svc:8080/v1/schema \
  -H 'Content-Type: application/json' -d @databases/weaviate/schema.json

# Neo4j
cypher-shell -u neo4j -p $NEO4J_PASSWORD < databases/neo4j/schema.cypher

# Dgraph
curl http://dgraph.databases.svc:8080/admin/schema --data-binary @databases/dgraph/schema.gql

# ScyllaDB
cqlsh scylladb.databases.svc -f databases/scylladb/schema.cql

# OpenSearch
curl -X PUT http://opensearch.databases.svc:9200/_index_template/logs \
  -H 'Content-Type: application/json' -d @databases/opensearch/logs-template.json

# TimescaleDB
psql -h timescaledb.databases.svc -U shopos -f databases/timescaledb/schema.sql
```

---

## Related

- HA Postgres (Patroni 3-node + PgBouncer): [`../infra/patroni/`](../infra/patroni/), [`../infra/pgbouncer/`](../infra/pgbouncer/)
- High-throughput Redis alternative: [`../infra/dragonfly/`](../infra/dragonfly/)
- Object storage: MinIO + SeaweedFS (in [`../helm/infra/`](../helm/infra/))
- Backup runbook (Postgres failover): [`../docs/runbooks/postgres-failover.md`](../docs/runbooks/postgres-failover.md)
