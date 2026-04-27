# ADR-002: Apache Kafka for All Asynchronous Cross-Domain Events

Status: Accepted  
Date: 2024-01-15  
Deciders: Platform Architecture Team

---

## Context

Many business operations trigger side effects across domain boundaries. When an order is placed, six or more services need to react: fulfillment-service reserves stock, loyalty-service accrues points, notification-orchestrator sends confirmation, analytics-service records the event, fraud-detection-service runs a scan, and accounting-service creates a journal entry.

Coupling these reactions synchronously into the checkout flow would increase latency by 500â€“2000ms and create hard dependency chains where a single downstream failure blocks checkout.

We evaluated: Apache Kafka, RabbitMQ, and direct async gRPC calls.

---

## Decision

Apache Kafka as the primary event bus for all asynchronous cross-domain communication.

- Topic naming convention: `{domain}.{entity}.{event}` â€” e.g., `commerce.order.placed`
- Event schemas are defined as Avro in `events/` with Schema Registry enforcing backward compatibility
- Debezium provides CDC from PostgreSQL and MongoDB into Kafka
- RabbitMQ handles task queues and delayed jobs within a single domain
- NATS JetStream handles real-time low-latency pub/sub (chat, presence, push notifications)

---

## Rationale

1. Durability â€” Kafka persists events to disk. Consumers that go down can replay from any offset.
2. Throughput â€” Handles millions of events/second; analytics and fraud detection process high-volume streams without backpressure on producers.
3. Decoupling â€” Producers have no knowledge of consumers. Adding a new consumer requires zero producer changes.
4. Event sourcing â€” The event-store-service uses Kafka as an append-only log, enabling full projection rebuilds and audit trails.
5. Stream processing â€” Apache Flink consumes Kafka topics for real-time fraud detection and analytics aggregations.

---

## Consequences

Positive: Zero producer-consumer coupling, complete audit trail via replay, real-time Flink processing, resilient to temporary consumer downtime.

Negative: Eventual consistency (not synchronous), operational overhead (ZooKeeper, Schema Registry, Kafka Connect), message ordering only per partition.

Mitigations: Kafka UI and AKHQ for topic browsing and lag monitoring; dead-letter-service for failed message handling; saga-orchestrator for multi-step distributed transactions requiring compensation.
