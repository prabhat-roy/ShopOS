# Communication Patterns — ShopOS

ShopOS uses six distinct communication patterns, each chosen for specific scenarios.

---

## Pattern Overview

```
─Œ───────────────────────────────────────────────────────────────────────────────
│                      ShopOS Communication Patterns                            │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  SYNCHRONOUS (gRPC / Protobuf)                                           │  │
│  │  Service A ──────────────────────────────────────────â–º Service B        │  │
│  │             immediate response required                                  │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  ASYNC EVENTS (Kafka / Avro)                                             │  │
│  │  Service A ──── event ──â–º Kafka topic ──â–º Service B, C, D (fan-out)     │  │
│  │             fire-and-forget cross-domain side effects                   │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  TASK QUEUES (RabbitMQ / AMQP)                                           │  │
│  │  Service A ──── task ──â–º Exchange ──â–º Worker (retry, DLQ)               │  │
│  │             background work within a domain, needs retries              │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  REAL-TIME (WebSocket / NATS JetStream)                                  │  │
│  │  Service â—„──── bidirectional ────â–º Client (browser/mobile)              │  │
│  │             live chat, in-app notifications, tracking updates           │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  CHANGE DATA CAPTURE (Debezium → Kafka)                                  │  │
│  │  Database ──── WAL/Oplog ──â–º Debezium ──â–º Kafka topic                   │  │
│  │             sync DB changes to search, analytics without app coupling   │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  DURABLE WORKFLOWS (Temporal)                                            │  │
│  │  Orchestrator ──── activity ──â–º Worker ──── result ──â–º Orchestrator     │  │
│  │             long-running sagas, guaranteed exactly-once execution        │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
│                                                                                │
│  ─Œ──────────────────────────────────────────────────────────────────────────  │
│  │  STREAM PROCESSING (Apache Flink)                                        │  │
│  │  Kafka ──── event stream ──â–º Flink job ──── aggregation ──â–º Kafka/DB    │  │
│  │             stateful real-time processing: fraud, analytics, enrichment  │  │
│  └──────────────────────────────────────────────────────────────────────────˜  │
└───────────────────────────────────────────────────────────────────────────────˜
```

---

## 1. Synchronous — gRPC (Protobuf)

Used for: Service-to-service calls that require an immediate response.

When to use: Reads and commands where the caller must wait for the result — for example, checkout calling payment, inventory, and tax in sequence before confirming an order.

```
checkout-service  ──gRPC──â–º  cart-service          GET cart contents
checkout-service  ──gRPC──â–º  inventory-service     reserve stock
checkout-service  ──gRPC──â–º  payment-service       process payment
checkout-service  ──gRPC──â–º  tax-service           calculate tax
checkout-service  ──gRPC──â–º  shipping-service      get rates
checkout-service  ──gRPC──â–º  promotions-service    apply discounts
checkout-service  ──gRPC──â–º  loyalty-service       apply points
```

Rules:
- All `.proto` files live in `proto/` at the repo root
- Port ranges allocated per domain (Platform: 50051–50059, Commerce: 50080–50099, etc.)
- All services implement `grpc.health.v1.Health`
- All gRPC clients use exponential backoff with jitter for transient failures
- mTLS enforced by Istio — no cleartext gRPC between pods

---

## 2. Asynchronous — Apache Kafka (Avro)

Used for: Cross-domain events where the producer must not wait for consumers.

When to use: A business event triggers reactions in multiple downstream domains — order placed → fulfil, notify, charge loyalty, record in accounting, run fraud scan.

Topic naming: `{domain}.{entity}.{event}`

```
commerce.order.placed
  ├──â–º fulfillment-service       reserve inventory, create shipment
  ├──â–º loyalty-service           accrue reward points
  ├──â–º notification-orchestrator send order confirmation
  ├──â–º analytics-service         record conversion event
  ├──â–º fraud-detection-service   post-purchase fraud scan
  └──â–º accounting-service        create journal entry

commerce.payment.processed
  ├──â–º invoice-service           generate PDF invoice
  ├──â–º accounting-service        record payment received
  └──â–º notification-orchestrator send payment receipt

identity.user.registered
  ├──â–º email-service             welcome email
  ├──â–º notification-orchestrator setup notification preferences
  └──â–º analytics-service         track new user acquisition

supplychain.shipment.updated
  ├──â–º notification-orchestrator send tracking update to customer
  └──â–º analytics-service         delivery performance analytics
```

Schema enforcement: All events are Avro schemas in `events/`. Confluent Schema Registry enforces backward compatibility — producers cannot break consumers.

Reliability: All consumers use consumer groups with explicit offset commits. Failed messages route to `dead-letter-service` after 3 retries.

---

## 3. Task Queues — RabbitMQ (AMQP)

Used for: Delayed jobs, retryable background tasks, and RPC-style patterns within a single domain.

When to use: Work that should run asynchronously but within a domain, requires reliable delivery with configurable retries, or needs scheduled/delayed execution.

```
scheduler-service  ──AMQP──â–º  worker-job-queue    scheduled cron job delivery
email-service      ──AMQP──â–º  smtp-worker          email delivery with retry
label-service      ──AMQP──â–º  print-queue           delayed label printing
```

Exchange types:

| Exchange | Use |
|---|---|
| `direct` | Point-to-point task delivery to a specific worker |
| `delayed` | Time-delayed execution (via RabbitMQ delayed message plugin) |
| `dead-letter` | Failed messages after max retries → `dead-letter-service` for inspection |

Retry policy: Up to 3 attempts with exponential backoff (5s → 25s → 125s). After third failure, message is moved to DLQ with full headers preserved for debugging.

---

## 4. Real-Time — WebSocket / NATS JetStream

Used for: Low-latency, bidirectional, real-time communication between services and end clients.

When to use: Live chat, in-app notifications, presence indicators, real-time order/shipment tracking.

```
live-chat-service           â—„──── WebSocket ────â–º browser / mobile client
in-app-notification-service â—„──── WebSocket ────â–º browser / mobile client
tracking-service            ──── NATS JetStream ──â–º mobile app (shipment updates)
push-notification-service   ──── NATS JetStream ──â–º FCM / APNs relay
```

NATS JetStream provides persistence and at-least-once delivery for real-time events, unlike core NATS which is fire-and-forget. JetStream consumers use pull-based subscription with acknowledgement.

---

## 5. Change Data Capture — Debezium → Kafka

Used for: Propagating database-level changes to downstream consumers without requiring application-level event emission.

When to use: Synchronising operational database state to search indexes, analytics stores, or read models without coupling the source service.

```
PostgreSQL orders table
  ──── WAL (Write-Ahead Log) ────â–º postgres-orders-connector (Debezium)
                                     └──â–º commerce.orders.cdc
                                            └──â–º ClickHouse (OLAP reporting)
                                            └──â–º OpenSearch (audit log search)

MongoDB catalog collection
  ──── Oplog ────â–º mongodb-catalog-connector (Debezium)
                     └──â–º catalog.products.cdc
                            └──â–º search-service (Elasticsearch index update)
                            └──â–º analytics-service (reporting sync)
```

Connectors configured:
- `postgres-orders-connector` — captures INSERT/UPDATE/DELETE on `orders` table
- `mongodb-catalog-connector` — captures product document changes

Consistency note: CDC consumers see changes in commit order. A consumer failure does not block the source service — Kafka durably buffers the change log.

---

## 6. Durable Workflows — Temporal

Used for: Long-running, multi-step business processes that must survive service restarts, network partitions, and partial failures.

When to use: Sagas spanning multiple services where each step must be retried independently, compensated on failure, or audited end-to-end.

```
Order Saga (Temporal Workflow)
  Step 1: reserve inventory     ──â–º inventory-service (gRPC)
  Step 2: calculate tax         ──â–º tax-service (gRPC)
  Step 3: charge payment        ──â–º payment-service (gRPC)
  Step 4: confirm order         ──â–º order-service (gRPC)
  Step 5: emit order.placed     ──â–º Kafka
  On failure at step 3:
    compensate step 1           ──â–º release inventory reservation

Subscription Renewal (Temporal Workflow)
  Schedule: daily
  Step 1: check renewal date
  Step 2: attempt charge        (retry up to 3 times over 7 days)
  Step 3: send notification
  Step 4: cancel on final failure
```

Why Temporal over saga-orchestrator alone:
- Persists workflow state in Temporal server — survives pod restarts mid-saga
- Built-in retry with configurable backoff per activity
- Full execution history available for audit and debugging
- `saga-orchestrator` handles simple choreography; Temporal handles complex orchestrated flows

---

## 7. Stream Processing — Apache Flink

Used for: Stateful, real-time processing of Kafka event streams — aggregations, enrichments, anomaly detection.

When to use: Analytics that require windowed aggregations, joining multiple streams, or processing that needs persistent state across millions of events.

```
Fraud Detection Job
  Kafka: commerce.order.placed + identity.login.failed
    ──â–º Flink (5-minute tumbling window, velocity checks)
         └──â–º security.fraud.detected ──â–º Kafka
               └──â–º fraud-detection-service (gRPC call to block order)

Order Analytics Job
  Kafka: commerce.order.placed + supplychain.shipment.updated
    ──â–º Flink (hourly revenue aggregation)
         └──â–º ClickHouse (orders_hourly materialized table)
         └──â–º analytics.revenue.aggregated ──â–º Kafka
```

State backend: RocksDB with checkpoint to S3/MinIO every 60 seconds. Exactly-once semantics via Kafka transactions + Flink checkpointing.

---

## Pattern Selection Guide

| Scenario | Pattern | Technology |
|---|---|---|
| Service A needs data from Service B immediately | Synchronous | gRPC |
| Business event triggers side effects in other domains | Async events | Kafka |
| Background job needs reliable retry logic | Task queue | RabbitMQ |
| Live updates pushed to browser/mobile client | Real-time | WebSocket / NATS |
| DB changes need to flow to other systems | CDC | Debezium → Kafka |
| Multi-step process spanning services, must be durable | Durable workflow | Temporal |
| Windowed aggregation or stream enrichment | Stream processing | Apache Flink |

---

## Failure Handling Summary

| Pattern | Retry strategy | Failure destination |
|---|---|---|
| gRPC | Exponential backoff + jitter (3 attempts) | Returns error to caller |
| Kafka consumers | 3 retries, then DLQ | `dead-letter-service` |
| RabbitMQ tasks | 3 attempts, exponential backoff | Dead letter exchange |
| Temporal activities | Per-activity retry policy (configurable) | Workflow compensation handler |
| Flink jobs | Automatic restart from last checkpoint | PagerDuty alert on repeated failure |

---

## References

- [ADR-001: gRPC](../adr/001-grpc-for-internal-communication.md)
- [ADR-002: Kafka](../adr/002-kafka-for-async-events.md)
- [Events / Avro schemas](../../events/)
- [Proto definitions](../../proto/)
- [Temporal config](../../workflow/temporal/)
- [Flink jobs](../../streaming/flink/)
- [Debezium connectors](../../streaming/debezium/)
