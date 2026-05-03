# ADR-005: Strict Database-per-Service, No Shared Data Stores

Status: Accepted  
Date: 2024-01-22  
Deciders: Platform Architecture Team, Data Architecture Lead

---

## Context

A common microservice anti-pattern is sharing a database between services — initially convenient, but it creates invisible coupling: one service's schema change can break another service's queries, and it becomes impossible to scale or migrate databases independently.

ShopOS has 154 services. We needed a clear rule on data ownership.

---

## Decision

Every service owns its own database schema. No two services share a database instance or schema. Cross-service data access is only permitted through the service's API (gRPC) or event stream (Kafka).

Database technology is chosen per service based on data shape and access patterns:

| Database | Version | When Used |
|---|---|---|
| PostgreSQL 16 | Primary relational store | Transactional data requiring ACID guarantees |
| MongoDB 8.0 | Document store | Flexible/nested documents (catalog, CMS, reviews, tracking) |
| Redis 7 | In-memory store | Cache, sessions, ephemeral state, pub/sub, rate limiting |
| Apache Cassandra 5.0 | Wide-column store | High-volume time-series events, analytics |
| Elasticsearch 8.15 | Search engine | Full-text search, faceted filtering |
| MinIO | Object storage | Binary files — images, videos, PDFs, exports |
| ClickHouse 24.8 | OLAP database | Reporting aggregations, analytics queries |
| Weaviate 1.26 | Vector database | Semantic search, ML embeddings |
| Neo4j 5.23 | Graph database | Product recommendations, relationship traversal |
| TimescaleDB 2.15 | Time-series (PostgreSQL extension) | Service metrics, inventory events, page views |
| Memcached 1.6 | High-throughput key-value cache | Hot read paths requiring maximum throughput, no persistence |

---

## Rationale

1. Independent deployability — A service can be migrated to a different database technology without affecting any other service.
2. Failure isolation — A database outage in one service does not cascade to unrelated services.
3. Independent scaling — High-traffic services (cart, session) use Redis. High-write analytics services use Cassandra. Each is sized for its own workload.
4. Schema autonomy — Teams own their schema migrations without coordinating with other teams.
5. Security — Compromising one service's credentials does not expose other services' data.

---

## Consequences

Positive: Full service autonomy; independent scaling and migration; failure isolation; schema changes are local.

Negative: No JOIN across service data; eventual consistency for aggregated views; increased operational complexity (11 different database technologies in production — Postgres, MongoDB, Redis, Memcached, Cassandra, Elasticsearch, OpenSearch, ClickHouse, Weaviate, Neo4j, TimescaleDB, MinIO).

Mitigations: The reporting-service and analytics-service use a read replica / data warehouse pattern to aggregate cross-domain data for reporting. Kafka CDC (Debezium) synchronises data between services when a denormalised read model is needed. Each service manages its own migrations (golang-migrate for Go, Flyway for Java/Kotlin).
