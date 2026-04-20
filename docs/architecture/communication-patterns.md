# Communication Patterns — ShopOS

ShopOS uses six distinct communication patterns, each chosen for specific scenarios.

---

## Pattern Overview

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                      ShopOS Communication Patterns                            │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  SYNCHRONOUS (gRPC / Protobuf)                                           │  │
│  │  Service A ──────────────────────────────────────────► Service B        │  │
│  │             immediate response required                                  │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  ASYNC EVENTS (Kafka / Avro)                                             │  │
│  │  Service A ──── event ──► Kafka topic ──► Service B, C, D (fan-out)     │  │
│  │             fire-and-forget cross-domain side effects                   │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  TASK QUEUES (RabbitMQ / AMQP)                                           │  │
│  │  Service A ──── task ──► Exchange ──► Worker (retry, DLQ)               │  │
│  │             background work within a domain, needs retries              │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  REAL-TIME (WebSocket / NATS JetStream)                                  │  │
│  │  Service ◄──── bidirectional ────► Client (browser/mobile)              │  │
│  │             live chat, in-app notifications, tracking updates           │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  CHANGE DATA CAPTURE (Debezium → Kafka)                                  │  │
│  │  Database ──── WAL/Oplog ──► Debezium ──► Kafka topic                   │  │
│  │             sync DB changes to search, analytics without app coupling   │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  DURABLE WORKFLOWS (Temporal)                                            │  │
│  │  Orchestrator ──── activity ──► Worker ──── result ──► Orchestrator     │  │
│  │             long-running sagas, guaranteed exactly-once execution        │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────┐  │
│  │  STREAM PROCESSING (Apache Flink)                                        │  │
│  │  Kafka ──── event stream ──► Flink job ──── aggregation ──► Kafka/DB    │  │
│  │             stateful real-time processing: fraud, analytics, enrichment  │  │
│  └─────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. Synchronous — gRPC (Protobuf)

**Used for:** Service-to-service calls that require an immediate response.

**When to use:** Reads and commands where the caller must wait for the result — for example, checkout calling payment, inventory, and tax in sequence before confirming an order.

```
checkout-service  ──gRPC──►  cart-service          GET cart contents
checkout-service  ──gRPC──►  inventory-service     reserve stock
checkout-service  ──gRPC──►  payment-service       process payment
checkout-service  ──gRPC──►  tax-service           calculate tax
checkout-service  ──gRPC──►  shipping-service      get rates
checkout-service  ──gRPC──►  promotions-service    apply discounts
checkout-service  ──gRPC──►  loyalty-service       apply points
```

**Rules:**
- All `.proto` files live in `proto/` at the repo root
- Port ranges allocated per domain (Platform: 50051–50059, Commerce: 50080–50099, etc.)
- All services implement `grpc.health.v1.Health`
- All gRPC clients use exponential backoff with jitter for transient failures
- mTLS enforced by Istio — no cleartext gRPC between pods

---

## 2. Asynchronous — Apache Kafka (Avro)

**Used for:** Cross-domain events where the producer must not wait for consumers.

**When to use:** A business event triggers reactions in multiple downstream domains — order placed → fulfil, notify, charge loyalty, record in accounting, run fraud scan.

**Topic naming:** `{domain}.{entity}.{event}`

```
commerce.order.placed
  ├──► fulfillment-service       reserve inventory, create shipment
  ├──► loyalty-service           accrue reward points
  ├──► notification-orchestrator send order confirmation
  ├──► analytics-service         record conversion event
  ├──► fraud-detection-service   post-purchase fraud scan
  └──► accounting-service        create journal entry

commerce.payment.processed
  ├──► invoice-service           generate PDF invoice
  ├──► accounting-service        record payment received
  └──► notification-orchestrator send payment receipt

identity.user.registered
  ├──► email-service             welcome email
  ├──► notification-orchestrator setup notification preferences
  └──► analytics-service         track new user acquisition

supplychain.shipment.updated
  ├──► notification-orchestrator send tracking update to customer
  └──► analytics-service         delivery performance analytics
```

**Schema enforcement:** All events are Avro schemas in `events/`. Confluent Schema Registry enforces backward compatibility — producers cannot break consumers.

**Reliability:** All consumers use consumer groups with explicit offset commits. Failed messages route to `dead-letter-service` after 3 retries.

---

## 3. Task Queues — RabbitMQ (AMQP)

**Used for:** Delayed jobs, retryable background tasks, and RPC-style patterns within a single domain.

**When to use:** Work that should run asynchronously but within a domain, requires reliable delivery with configurable retries, or needs scheduled/delayed execution.

```
scheduler-service  ──AMQP──►  worker-job-queue    scheduled cron job delivery
email-service      ──AMQP──►  smtp-worker          email delivery with retry
label-service      ──AMQP──►  print-queue           delayed label printing
```

**Exchange types:**

| Exchange | Use |
|---|---|
| `direct` | Point-to-point task delivery to a specific worker |
| `delayed` | Time-delayed execution (via RabbitMQ delayed message plugin) |
| `dead-letter` | Failed messages after max retries → `dead-letter-service` for inspection |

**Retry policy:** Up to 3 attempts with exponential backoff (5s → 25s → 125s). After third failure, message is moved to DLQ with full headers preserved for debugging.

---

## 4. Real-Time — WebSocket / NATS JetStream

**Used for:** Low-latency, bidirectional, real-time communication between services and end clients.

**When to use:** Live chat, in-app notifications, presence indicators, real-time order/shipment tracking.

```
live-chat-service           ◄──── WebSocket ────► browser / mobile client
in-app-notification-service ◄──── WebSocket ────► browser / mobile client
tracking-service            ──── NATS JetStream ──► mobile app (shipment updates)
push-notification-service   ──── NATS JetStream ──► FCM / APNs relay
```

**NATS JetStream** provides persistence and at-least-once delivery for real-time events, unlike core NATS which is fire-and-forget. JetStream consumers use pull-based subscription with acknowledgement.

---

## 5. Change Data Capture — Debezium → Kafka

**Used for:** Propagating database-level changes to downstream consumers without requiring application-level event emission.

**When to use:** Synchronising operational database state to search indexes, analytics stores, or read models without coupling the source service.

```
PostgreSQL orders table
  ──── WAL (Write-Ahead Log) ────► postgres-orders-connector (Debezium)
                                     └──► commerce.orders.cdc
                                            └──► ClickHouse (OLAP reporting)
                                            └──► OpenSearch (audit log search)

MongoDB catalog collection
  ──── Oplog ────► mongodb-catalog-connector (Debezium)
                     └──► catalog.products.cdc
                            └──► search-service (Elasticsearch index update)
                            └──► analytics-service (reporting sync)
```

**Connectors configured:**
- `postgres-orders-connector` — captures INSERT/UPDATE/DELETE on `orders` table
- `mongodb-catalog-connector` — captures product document changes

**Consistency note:** CDC consumers see changes in commit order. A consumer failure does not block the source service — Kafka durably buffers the change log.

---

## 6. Durable Workflows — Temporal

**Used for:** Long-running, multi-step business processes that must survive service restarts, network partitions, and partial failures.

**When to use:** Sagas spanning multiple services where each step must be retried independently, compensated on failure, or audited end-to-end.

```
Order Saga (Temporal Workflow)
  Step 1: reserve inventory     ──► inventory-service (gRPC)
  Step 2: calculate tax         ──► tax-service (gRPC)
  Step 3: charge payment        ──► payment-service (gRPC)
  Step 4: confirm order         ──► order-service (gRPC)
  Step 5: emit order.placed     ──► Kafka
  On failure at step 3:
    compensate step 1           ──► release inventory reservation

Subscription Renewal (Temporal Workflow)
  Schedule: daily
  Step 1: check renewal date
  Step 2: attempt charge        (retry up to 3 times over 7 days)
  Step 3: send notification
  Step 4: cancel on final failure
```

**Why Temporal over saga-orchestrator alone:**
- Persists workflow state in Temporal server — survives pod restarts mid-saga
- Built-in retry with configurable backoff per activity
- Full execution history available for audit and debugging
- `saga-orchestrator` handles simple choreography; Temporal handles complex orchestrated flows

---

## 7. Stream Processing — Apache Flink

**Used for:** Stateful, real-time processing of Kafka event streams — aggregations, enrichments, anomaly detection.

**When to use:** Analytics that require windowed aggregations, joining multiple streams, or processing that needs persistent state across millions of events.

```
Fraud Detection Job
  Kafka: commerce.order.placed + identity.login.failed
    ──► Flink (5-minute tumbling window, velocity checks)
         └──► security.fraud.detected ──► Kafka
               └──► fraud-detection-service (gRPC call to block order)

Order Analytics Job
  Kafka: commerce.order.placed + supplychain.shipment.updated
    ──► Flink (hourly revenue aggregation)
         └──► ClickHouse (orders_hourly materialized table)
         └──► analytics.revenue.aggregated ──► Kafka
```

**State backend:** RocksDB with checkpoint to S3/MinIO every 60 seconds. Exactly-once semantics via Kafka transactions + Flink checkpointing.

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
