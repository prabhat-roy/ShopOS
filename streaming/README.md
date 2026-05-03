# Streaming — ShopOS

Change Data Capture and real-time stream processing. Debezium captures database mutations
and streams them into Kafka; Apache Flink consumes those streams for real-time analytics
and fraud detection.

---

## Directory Structure

```
streaming/
├── debezium/
│   ├── postgres-connector.json     ← CDC connector for PostgreSQL (order, payment, user tables)
│   └── mongodb-connector.json      ← CDC connector for MongoDB (product catalog, reviews)
└── flink/
    ├── order-analytics.yaml        ← FlinkDeployment — real-time order aggregations
    └── fraud-detection.yaml        ← FlinkDeployment — streaming fraud scoring
```

---

## Debezium — Change Data Capture

Debezium is deployed via the Kafka Connect framework and captures row-level changes from
databases into Kafka topics in real time.

### PostgreSQL Connector

Captures changes from the `orders`, `payments`, and `users` tables in the `shopos` database.

Output topics:

| Topic | Source table | Description |
|---|---|---|
| `dbz.public.orders` | `orders` | INSERT/UPDATE/DELETE on orders |
| `dbz.public.payments` | `payments` | Payment state changes |
| `dbz.public.users` | `users` | User profile updates |

Deploy:
```bash
curl -X POST http://kafka-connect.messaging.svc:8083/connectors \
  -H 'Content-Type: application/json' \
  -d @streaming/debezium/postgres-connector.json
```

Check status:
```bash
curl http://kafka-connect.messaging.svc:8083/connectors/postgres-cdc/status
```

### MongoDB Connector

Captures change streams from MongoDB collections: `products`, `reviews`, `cms_pages`.

Output topics: `dbz.catalog.products`, `dbz.catalog.reviews`, `dbz.content.cms_pages`

Deploy:
```bash
curl -X POST http://kafka-connect.messaging.svc:8083/connectors \
  -H 'Content-Type: application/json' \
  -d @streaming/debezium/mongodb-connector.json
```

---

## Apache Flink

Flink jobs run as `FlinkDeployment` custom resources managed by the Flink Kubernetes Operator.

### Order Analytics Job

Consumes `commerce.order.placed` and `dbz.public.orders` topics. Computes:
- Rolling 5-minute order volume per region
- Real-time revenue by category
- Running totals for the operations dashboard

Deploy:
```bash
kubectl apply -f streaming/flink/order-analytics.yaml -n analytics-ai
```

Check job status:
```bash
kubectl get flinkdeployment order-analytics -n analytics-ai
```

### Fraud Detection Job

Consumes `commerce.payment.processed` and enriches with user behaviour signals from
`analytics.page.viewed`. Emits `security.fraud.detected` events when the score exceeds
the configured threshold.

Deploy:
```bash
kubectl apply -f streaming/flink/fraud-detection.yaml -n analytics-ai
```

---

## Data Flow

```
PostgreSQL ──â–º Debezium ──â–º Kafka (dbz.public.*) ──â–º Flink (order-analytics) ──â–º ClickHouse
MongoDB    ──â–º Debezium ──â–º Kafka (dbz.catalog.*)
                                                  ──â–º Flink (fraud-detection) ──â–º security.fraud.detected
```

---

## Prerequisites

| Component | Namespace | Required by |
|---|---|---|
| Kafka + ZooKeeper | `messaging` | Debezium, Flink |
| Kafka Connect | `messaging` | Debezium connectors |
| Schema Registry | `messaging` | Avro serialisation |
| Flink Kubernetes Operator | `analytics-ai` | FlinkDeployment CRDs |
| ClickHouse | `analytics-ai` | Flink output sink |

---

## References

- [Debezium Documentation](https://debezium.io/documentation/)
- [Apache Flink on Kubernetes](https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-main/)
- [Kafka topics reference](../events/README.md)
- [ClickHouse schema](../databases/clickhouse/)
