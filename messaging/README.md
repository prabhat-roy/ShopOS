# Messaging — ShopOS

Configuration and Helm charts for all messaging infrastructure components.

## Directory Structure

```
messaging/
├── kafka/              ← Confluent Kafka broker configs (Kraft mode)
├── kafka-connect/      ← Kafka Connect worker configurations
├── kafka-ui/           ← Kafka UI (Provectus) deployment
├── akhq/               ← AKHQ Kafka management UI
├── schema-registry/    ← Confluent Schema Registry (Avro enforcement)
├── ksqldb/             ← ksqlDB stream processing
├── zookeeper/          ← ZooKeeper (Kafka coordination)
├── rabbitmq/           ← RabbitMQ 3.13 (task queues, delayed messages)
├── nats/               ← NATS JetStream 2.10 (real-time low-latency)
├── strimzi/            ← Strimzi Kafka Operator for Kubernetes
├── redpanda/           ← Redpanda (Kafka-compatible alternative)
├── pulsar/             ← Apache Pulsar (alternative event streaming)
├── memphis/            ← Memphis.dev (developer-friendly message broker)
└── activemq-artemis/   ← ActiveMQ Artemis (JMS-compatible broker)
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| **Kafka** (Confluent) | 7.7.1 | Primary event streaming — domain events |
| **Schema Registry** | 7.7.1 | Avro schema enforcement and versioning |
| **ZooKeeper** | 7.7.1 | Kafka coordination (pre-KRaft clusters) |
| **RabbitMQ** | 3.13 | Task queues, delayed jobs, dead-letter routing |
| **NATS JetStream** | 2.10 | Real-time pub/sub (chat, notifications, tracking) |
| **Kafka UI** | latest | Web UI for topic browsing and consumer group monitoring |

## Usage Pattern

| Use Case | Broker | Example |
|---|---|---|
| Cross-domain business events | Kafka | `commerce.order.placed` → fulfilment, loyalty, analytics |
| Background jobs with retry | RabbitMQ | Email delivery, label printing, scheduled reports |
| Real-time client updates | NATS | Live chat, shipment tracking, in-app notifications |
| CDC from databases | Kafka Connect + Debezium | Postgres/MongoDB → Kafka topics |

## Kafka Topic Naming

`{domain}.{entity}.{event}` — e.g., `commerce.order.placed`, `identity.user.registered`

All Avro schemas are in `events/` at the repo root. Schema Registry enforces backward compatibility.

## RabbitMQ Exchange Types

- `direct` — point-to-point task delivery
- `x-delayed-message` — delayed job execution (requires delayed-message plugin)
- Dead-letter exchange — failed messages routed to `dead-letter-service`

## References

- [Communication Patterns](../docs/architecture/communication-patterns.md)
- [Avro Event Schemas](../events/README.md)
- [ADR-002: Kafka for Async Events](../docs/adr/002-kafka-for-async-events.md)
