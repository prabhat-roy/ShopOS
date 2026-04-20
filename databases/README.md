# Databases — ShopOS

Specialist database schemas for OLAP, vector search, graph analytics, high-throughput
time-series, and log analytics. These complement the primary transactional databases
(PostgreSQL, MongoDB, Redis) that are configured per-service via Docker Compose and Helm.

---

## Directory Structure

```
databases/
├── clickhouse/         ← OLAP schema — orders, events, revenue_daily MV
├── weaviate/           ← Vector schema — Product, UserQuery classes
├── neo4j/              ← Graph schema — product recommendation graph
├── scylladb/           ← High-throughput time-series keyspace
└── opensearch/         ← Index templates + ILM policies
```

---

## ClickHouse

**Role:** OLAP analytics — order aggregations, revenue reporting, funnel analysis.

**Key objects:**

| Object | Type | Description |
|---|---|---|
| `orders` | Table | ReplicatedMergeTree — raw order rows |
| `order_events` | Table | ReplicatedMergeTree — state-change events |
| `revenue_daily` | Materialized View | Aggregates daily revenue from `orders` |
| `funnel_events` | Table | User funnel step tracking |

Connect:
```bash
clickhouse-client --host clickhouse.analytics-ai.svc --port 9000 --user shopos
```

---

## Weaviate

**Role:** Vector database for semantic product search and RAG (retrieval-augmented generation).

**Classes:**

| Class | Vectorizer | Description |
|---|---|---|
| `Product` | `text2vec-openai` | Product name, description, category embeddings |
| `UserQuery` | `text2vec-openai` | Historical user search queries for personalisation |

REST client:
```python
import weaviate
client = weaviate.Client("http://weaviate.analytics-ai.svc:8080")
results = client.query.get("Product", ["name", "description"]) \
    .with_near_text({"concepts": ["wireless headphones"]}) \
    .with_limit(10).do()
```

---

## Neo4j

**Role:** Graph database powering the product recommendation engine.

**Graph model:**

```
(User)-[:VIEWED]->(Product)
(User)-[:PURCHASED]->(Product)
(Product)-[:BELONGS_TO]->(Category)
(Product)-[:FREQUENTLY_BOUGHT_WITH]->(Product)
```

Cypher — "users who bought X also bought":
```cypher
MATCH (p:Product {id: $productId})<-[:PURCHASED]-(u:User)-[:PURCHASED]->(rec:Product)
WHERE rec.id <> $productId
RETURN rec, count(u) AS coOccurrence
ORDER BY coOccurrence DESC LIMIT 10
```

---

## ScyllaDB

**Role:** High-throughput Cassandra-compatible time-series store for analytics events,
session data, and IoT-style telemetry from the supply chain.

**Keyspace:** `shopos_analytics`

**Key tables:**

| Table | Partition Key | Clustering Key | Description |
|---|---|---|---|
| `page_views` | `(user_id, date)` | `timestamp` | Page view events |
| `product_clicks` | `(product_id, date)` | `timestamp` | Product click events |
| `shipment_telemetry` | `(shipment_id)` | `recorded_at` | Cold-chain sensor readings |

Replication: `NetworkTopologyStrategy`, RF=3.

Connect:
```bash
cqlsh scylladb.messaging.svc 9042 -u shopos
```

---

## OpenSearch

**Role:** Log analytics, audit trail, and security event search. Alternative to the ELK stack.

**Index templates:**

| Template | Applies to | Shards | Replicas | Retention |
|---|---|---|---|---|
| `logs-*` | Application logs | 3 | 1 | 30 days |
| `audit-*` | Audit events | 2 | 1 | 365 days |
| `security-*` | Falco / security events | 2 | 1 | 90 days |

**ILM policies:**

| Policy | Hot phase | Warm phase | Delete |
|---|---|---|---|
| `logs-policy` | 7 days | 23 days | 30 days |
| `audit-policy` | 30 days | 335 days | 365 days |

---

## Applying Schemas

```bash
# ClickHouse
clickhouse-client --host <host> < databases/clickhouse/schema.sql

# Weaviate (via REST)
curl -X POST http://weaviate:8080/v1/schema \
  -H 'Content-Type: application/json' \
  -d @databases/weaviate/schema.json

# Neo4j (via Cypher shell)
cypher-shell -u neo4j -p <pass> < databases/neo4j/schema.cypher

# ScyllaDB
cqlsh <host> -f databases/scylladb/schema.cql

# OpenSearch index templates
curl -X PUT http://opensearch:9200/_index_template/logs \
  -H 'Content-Type: application/json' \
  -d @databases/opensearch/logs-template.json
```
