# Database Strategy — ShopOS

ShopOS uses **database-per-service** with technology chosen per access pattern. No two services share a database instance or schema. Cross-service data access always goes through gRPC or Kafka — never via direct DB connection.

---

## Technology Selection Matrix

| Database | Version | Access Pattern | Services |
|---|---|---|---|
| **PostgreSQL** | 16 | ACID transactions, relational queries, strong consistency | 60+ services (orders, payments, users, inventory, financial) |
| **MongoDB** | 8.0 | Flexible schema, nested documents, rapid iteration | product-catalog, cms-service, review-rating, tracking, cold-chain |
| **Redis** | 7 | Sub-millisecond reads, ephemeral data, pub/sub, sorted sets | cart, sessions, rate-limiter, waitlist, flash-sale, live-chat |
| **Memcached** | 1.6 | High-throughput simple key-value cache (no persistence needed) | Hot read paths requiring maximum throughput with minimal overhead |
| **Cassandra** | 5.0 | High-write time-series, wide rows, no single point of failure | analytics-service, event-tracking, data-pipeline, reporting |
| **Elasticsearch** | 8.15 | Full-text search, faceted filtering, relevance ranking | search-service |
| **OpenSearch** | 2.17 | Log analytics, audit log search, alternative product search | Log pipeline, audit search, OpenSearch-based deployments |
| **ClickHouse** | 24.8 | OLAP, columnar storage, aggregations over billions of rows | reporting-service, analytics aggregations, revenue dashboards |
| **Weaviate** | 1.26 | Vector embeddings, semantic similarity, ANN search | recommendation-service, semantic product search |
| **Neo4j** | 5.23 | Graph traversal, relationship queries | recommendation-service (product graph, collaborative filtering) |
| **MinIO** | latest | S3-compatible object storage for large binary files | media-asset, document-service, video-service, data-export |

---

## Decision Rules

### Use PostgreSQL when:
- Data has relationships requiring JOINs (orders → order_items → products)
- ACID transactions are required (payment processing, inventory reservation, financial ledger)
- Strong consistency matters more than write throughput
- Data is regulated (financial, PII) — PostgreSQL row-level encryption + Vault Transit available
- Team needs standard SQL tooling and Flyway/golang-migrate migrations

### Use MongoDB when:
- Schema evolves frequently (product attributes vary by category)
- Data is naturally hierarchical or nested (product variants, CMS page content, tracking events)
- Horizontal scaling is needed for document reads
- Fields differ significantly across documents (configurator data, review metadata)

### Use Redis when:
- Data has a natural TTL (sessions expire, cart items cleared after checkout, flash-sale windows)
- Sub-millisecond latency is required (rate limiting, feature flags, leaderboards)
- Data can be rebuilt from the source of truth if lost (cache invalidation acceptable)
- Real-time pub/sub is needed within a single service

### Use Memcached when:
- Cache entries are simple string or binary values (no complex data types)
- Horizontal scalability via consistent hashing is needed across many nodes
- Maximum raw throughput matters and Redis features (TTL per key, pub/sub, Lua) are not needed

### Use Cassandra when:
- Write throughput is >10k/sec sustained
- Data is time-series or append-only (events, logs, metrics, click streams)
- Queries are always by partition key (no ad-hoc queries or JOINs needed)
- Multi-region replication with tunable consistency is required

### Use Elasticsearch when:
- Full-text search with relevance ranking is a core use case
- Faceted filtering (by brand, price range, category, rating) is required
- Near-real-time index updates are needed (Debezium CDC feeds product changes)
- Geo-search or percolation queries are needed

### Use ClickHouse when:
- Queries aggregate over hundreds of millions of rows
- Data is write-once analytical (order history, revenue, event rollups)
- Materialized views with incremental refresh are needed (e.g., `revenue_daily`)
- Columnar compression is needed to reduce storage cost for analytics data

### Use Weaviate when:
- Semantic similarity search over embeddings is required
- ML model outputs (vectors) need to be queried by nearest neighbours
- Combining vector search with structured filters (e.g., category + semantic)

### Use Neo4j when:
- Relationship traversal is the primary query pattern (product recommendations via co-purchase graph)
- Data is a naturally connected graph (users → purchased → products → belong_to → categories)
- Multi-hop queries would require expensive self-JOINs in SQL

---

## Migration Strategy

Each service manages its own migrations:

| Language | Migration tool | Location |
|---|---|---|
| Go | `golang-migrate` | `db/migrations/*.sql` |
| Java | Flyway | `src/main/resources/db/migration/V*.sql` |
| Kotlin | Flyway | `src/main/resources/db/migration/V*.sql` |
| Python | Alembic | `migrations/versions/` |
| Node.js | Knex / Prisma | `migrations/` |
| C# | EF Core Migrations | `Migrations/` |

Migrations run automatically on service startup. Both Flyway and `golang-migrate` are idempotent — re-running an already-applied migration is a no-op. All migration files are checked into the service's source tree and applied in order by version number.

---

## Read Model Strategy

Operational databases are never queried directly by the reporting or analytics layer.

```
Operational DB (PostgreSQL / MongoDB)
  └──► Debezium CDC
         └──► Kafka topic
                ├──► ClickHouse (OLAP read model — orders, revenue, inventory)
                ├──► OpenSearch (log and audit read model)
                └──► Elasticsearch (product search index)
```

This ensures:
- Analytical queries never add load to transactional databases
- Read models can be rebuilt from Kafka topic replay if a store is corrupted
- The reporting layer has no knowledge of operational DB schemas

---

## Data Access Rules

1. A service **MUST NOT** connect to another service's database directly
2. A service **MUST NOT** import another service's ORM models or schema types
3. Cross-service data access **MUST** go through gRPC (synchronous) or Kafka (asynchronous)
4. Read models for reporting are populated via Kafka CDC, not by querying operational databases
5. All database credentials are issued dynamically by **HashiCorp Vault** (short-lived, rotated per pod)
6. All PII fields are encrypted at rest using **Vault Transit** engine before storage

---

## Backup & Recovery

| Store | Backup method | RPO | RTO |
|---|---|---|---|
| PostgreSQL | Velero + WAL archiving to MinIO/S3 | 1h | 30m |
| MongoDB | Velero + mongodump to MinIO/S3 | 1h | 30m |
| Redis | AOF persistence + Velero snapshot | 15m | 5m |
| Cassandra | nodetool snapshot → MinIO/S3 | 4h | 1h |
| Elasticsearch | snapshot API → MinIO/S3 | 1h | 30m |
| ClickHouse | BACKUP TO S3 (native) | 4h | 1h |
| MinIO | bucket replication to secondary MinIO | 5m | 5m |

Velero runs a daily cluster-wide backup at 02:00 UTC (`kubernetes/velero/`). Individual database snapshots run on their own schedules.

---

## References

- [ADR-005: Database-per-Service](../adr/005-database-per-service.md)
- [Communication Patterns](communication-patterns.md)
- [Streaming / Debezium](../../streaming/debezium/)
- [ClickHouse schema](../../databases/clickhouse/)
- [Weaviate schema](../../databases/weaviate/)
- [Neo4j schema](../../databases/neo4j/)
- [TimescaleDB schema](../../databases/timescaledb/)
- [OpenSearch schema](../../databases/opensearch/)
