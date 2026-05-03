# Messaging — ShopOS

All async/streaming infrastructure: Kafka, low-latency pub/sub, change-data-capture, broker
governance, and protocol mediation.

## Layout

```
messaging/
├── kafka/                 Strimzi Kafka cluster + KafkaTopic CRDs (topics.yaml — 20 Avro topics + 3 DLQs)
├── kafka-connect/         Kafka Connect workers (Debezium connectors live in streaming/debezium/)
├── kafka-ui/              Provectus Kafka UI
├── akhq/                  AKHQ Kafka management UI
├── kafka-monitor/         End-to-end latency + availability monitoring
├── schema-registry/       Confluent Schema Registry — Avro contract enforcement
├── ksqldb/                ksqlDB stream processing
├── zookeeper/             ZooKeeper (Kafka coordination — pre-KRaft)
├── strimzi/               Strimzi Kafka Operator for Kubernetes
├── conduktor/             Conduktor Gateway interceptors — schema validation, rate limiting,
│                          PII masking, audit log on all Kafka topics
├── rabbitmq/              RabbitMQ 3.13 — task queues, delayed messages, RPC
├── nats/                  NATS JetStream 2.10 — real-time pub/sub
├── redpanda/              Redpanda 5.9 — Kafka-API-compatible low-latency alternative
├── zilla/                 Zilla — Kafka → REST/SSE/MQTT proxy for browsers/IoT
├── pulsar/                Apache Pulsar (alternative event streaming)
├── memphis/               Memphis.dev (developer-friendly broker)
└── activemq-artemis/      ActiveMQ Artemis (JMS)
```

## Deployed stack

| Component | Version | Role |
|---|---|---|
| Apache Kafka (Strimzi) | 7.7.1 | Primary event streaming — domain events |
| `KafkaTopic` CRDs | n/a | 20 Avro topics + 3 DLQs in [`kafka/topics.yaml`](kafka/topics.yaml) — provisioned at install |
| Schema Registry | 7.7.1 | Avro schema enforcement |
| Conduktor Gateway | 3.3 | Policy proxy (schema, rate-limit, PII mask, audit) |
| RabbitMQ | 3.13 | Task queues, delayed jobs, dead-letter routing |
| NATS JetStream | 2.10 | Real-time pub/sub (chat, notifications, presence) |
| Redpanda | 5.9 | Low-latency Kafka-compatible alternative for analytics |
| Zilla | 0.9 | Browser/IoT-friendly protocol mediation over Kafka |

## Topic naming

`{domain}.{entity}.{event}` — e.g. `commerce.order.placed`, `identity.user.registered`.
Avro schemas live in [`../events/`](../events/) at the repo root. Schema Registry enforces
backward compatibility on every register.

DLQ topics use the `dlq.<group>` prefix: `dlq.commerce.order`, `dlq.notification`, `dlq.analytics`.

## Provisioning topics

```bash
kubectl apply -f messaging/kafka/topics.yaml
# Verify
kubectl get kafkatopics -n streaming
```

## Use-case → broker

| Use case | Broker | Example |
|---|---|---|
| Cross-domain business events | Kafka | `commerce.order.placed` → fulfilment, loyalty, analytics |
| Background jobs with retry | RabbitMQ | Email delivery, label printing, scheduled reports |
| Real-time client updates | NATS JetStream | Live chat, shipment tracking, in-app notifications |
| CDC from databases | Debezium → Kafka | Postgres/MongoDB → Kafka topics |
| Browser-side streaming | Zilla → Kafka | Storefront live event feed via SSE |
| Analytics-ai low-latency | Redpanda | High-throughput event collection without JVM overhead |
| Kafka governance | Conduktor Gateway | Schema enforcement, rate limit, PII mask, audit |

## References

- [Avro event schemas](../events/README.md)
- [ADR-002: Kafka for Async Events](../docs/adr/002-kafka-for-async-events.md)
- [Kafka consumer lag runbook](../docs/runbooks/kafka-consumer-lag.md)
- [Streaming (Debezium + Flink)](../streaming/README.md)
