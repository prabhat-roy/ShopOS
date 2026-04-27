# Messaging â€” ShopOS

Configuration and Helm charts for all messaging infrastructure components.

## Directory Structure

```
messaging/
â”œâ”€â”€ kafka/              â† Confluent Kafka broker configs (Kraft mode)
â”œâ”€â”€ kafka-connect/      â† Kafka Connect worker configurations
â”œâ”€â”€ kafka-ui/           â† Kafka UI (Provectus) deployment
â”œâ”€â”€ akhq/               â† AKHQ Kafka management UI
â”œâ”€â”€ schema-registry/    â† Confluent Schema Registry (Avro enforcement)
â”œâ”€â”€ ksqldb/             â† ksqlDB stream processing
â”œâ”€â”€ zookeeper/          â† ZooKeeper (Kafka coordination)
â”œâ”€â”€ rabbitmq/           â† RabbitMQ 3.13 (task queues, delayed messages)
â”œâ”€â”€ nats/               â† NATS JetStream 2.10 (real-time low-latency)
â”œâ”€â”€ strimzi/            â† Strimzi Kafka Operator for Kubernetes
â”œâ”€â”€ redpanda/           â† Redpanda (Kafka-compatible alternative)
â”œâ”€â”€ pulsar/             â† Apache Pulsar (alternative event streaming)
â”œâ”€â”€ memphis/            â† Memphis.dev (developer-friendly message broker)
â””â”€â”€ activemq-artemis/   â† ActiveMQ Artemis (JMS-compatible broker)
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| Kafka (Confluent) | 7.7.1 | Primary event streaming â€” domain events |
| Schema Registry | 7.7.1 | Avro schema enforcement and versioning |
| ZooKeeper | 7.7.1 | Kafka coordination (pre-KRaft clusters) |
| RabbitMQ | 3.13 | Task queues, delayed jobs, dead-letter routing |
| NATS JetStream | 2.10 | Real-time pub/sub (chat, notifications, tracking) |
| Kafka UI | latest | Web UI for topic browsing and consumer group monitoring |

## Usage Pattern

| Use Case | Broker | Example |
|---|---|---|
| Cross-domain business events | Kafka | `commerce.order.placed` â†’ fulfilment, loyalty, analytics |
| Background jobs with retry | RabbitMQ | Email delivery, label printing, scheduled reports |
| Real-time client updates | NATS | Live chat, shipment tracking, in-app notifications |
| CDC from databases | Kafka Connect + Debezium | Postgres/MongoDB â†’ Kafka topics |

## Kafka Topic Naming

`{domain}.{entity}.{event}` â€” e.g., `commerce.order.placed`, `identity.user.registered`

All Avro schemas are in `events/` at the repo root. Schema Registry enforces backward compatibility.

## RabbitMQ Exchange Types

- `direct` â€” point-to-point task delivery
- `x-delayed-message` â€” delayed job execution (requires delayed-message plugin)
- Dead-letter exchange â€” failed messages routed to `dead-letter-service`

## References

- [Communication Patterns](../docs/architecture/communication-patterns.md)
- [Avro Event Schemas](../events/README.md)
- [ADR-002: Kafka for Async Events](../docs/adr/002-kafka-for-async-events.md)
