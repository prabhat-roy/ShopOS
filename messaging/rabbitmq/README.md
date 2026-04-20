# RabbitMQ — ShopOS Messaging Layer

## Role in ShopOS

RabbitMQ serves as the **AMQP-based task queue and RPC backbone** in ShopOS, complementing
Apache Kafka (which handles high-throughput event streaming). While Kafka is the canonical
event bus for domain events and analytics, RabbitMQ covers:

| Use Case | Details |
|---|---|
| **Task queues** | Discrete units of work dispatched to a worker pool (e.g. background jobs, report generation, image resizing) |
| **Delayed / scheduled messages** | `x-delayed-message` plugin — retry-after-delay, scheduled notifications |
| **RPC over AMQP** | Request-reply pattern for synchronous-feeling async calls (reply-to + correlation-id) |
| **Fan-out notifications** | Single publish → multiple notification channels (email, SMS, push) via fanout exchange |
| **Dead-letter handling** | Automatic routing of failed messages for inspection and manual replay |
| **Priority queues** | `worker.jobs` supports 10 priority levels for urgent vs. batch work |

---

## Exchange / Queue Topology

```mermaid
graph LR
    subgraph Producers
        CO[checkout-service]
        PS[payment-service]
        NO[notification-orchestrator]
        WQ[worker-job-queue]
    end

    subgraph Exchanges
        EO["commerce.orders\n(topic)"]
        EP["commerce.payments\n(topic)"]
        EN["notifications\n(fanout)"]
        EW["worker.tasks\n(direct)"]
        EDL["dead-letter\n(direct)"]
    end

    subgraph Queues
        QO["order.processing"]
        QP["payment.processing"]
        QE["notification.email"]
        QS["notification.sms"]
        QPU["notification.push"]
        QW["worker.jobs"]
        QDL["dead-letter.queue"]
    end

    subgraph Consumers
        OS[order-service]
        PAY[payment-service]
        EML[email-service]
        SMS[sms-service]
        PNS[push-notification-service]
        WJS[worker-job-queue service]
        OPS[ops / manual replay]
    end

    CO -->|"order.placed"| EO
    EO -->|"order.#"| QO
    QO --> OS

    PS -->|"payment.initiated"| EP
    EP -->|"payment.#"| QP
    QP --> PAY

    NO --> EN
    EN --> QE --> EML
    EN --> QS --> SMS
    EN --> QPU --> PNS

    WQ -->|"job"| EW
    EW -->|"job"| QW --> WJS

    QO -. "nack / expire" .-> EDL
    QP -. "nack / expire" .-> EDL
    QE -. "nack / expire" .-> EDL
    QS -. "nack / expire" .-> EDL
    QPU -. "nack / expire" .-> EDL
    QW -. "nack / expire" .-> EDL
    EDL -->|"dead-letter.queue"| QDL --> OPS
```

---

## Exchange Reference

| Exchange | Type | Routing Key Pattern | Purpose |
|---|---|---|---|
| `commerce.orders` | topic | `order.*` | Order lifecycle events |
| `commerce.payments` | topic | `payment.*` | Payment events |
| `notifications` | fanout | — | Broadcast to all notification channels |
| `dead-letter` | direct | `dead-letter.queue` | Catch undeliverable / expired messages |
| `commerce.orders.delayed` | x-delayed-message/topic | `order.*` | Retry-after-delay, scheduled reminders |
| `worker.tasks` | direct | `job` | Dispatch background jobs |

## Queue Reference

| Queue | Exchange | TTL | Max Length | Purpose |
|---|---|---|---|---|
| `order.processing` | `commerce.orders` | 24 h | 100,000 | New orders from checkout |
| `payment.processing` | `commerce.payments` | 1 h | 50,000 | Payment commands |
| `notification.email` | `notifications` | 24 h | 500,000 | Outbound emails |
| `notification.sms` | `notifications` | 1 h | 100,000 | SMS messages |
| `notification.push` | `notifications` | 1 h | 200,000 | Push notifications |
| `worker.jobs` | `worker.tasks` | 7 d | 1,000,000 | Background jobs (priority 0–10) |
| `dead-letter.queue` | `dead-letter` | 30 d | 100,000 | Failed messages |

---

## Usage Examples

### Publishing an Order Event (Go)

```go
import amqp "github.com/rabbitmq/amqp091-go"

conn, _ := amqp.Dial("amqp://shopos-app:shopos-app@rabbitmq:5672/shopos")
ch, _   := conn.Channel()

body, _ := json.Marshal(OrderPlacedEvent{OrderID: "ord-123", Total: 99.99})
ch.Publish(
    "commerce.orders", // exchange
    "order.placed",    // routing key
    false,             // mandatory
    false,             // immediate
    amqp.Publishing{
        ContentType:  "application/json",
        DeliveryMode: amqp.Persistent,
        Body:         body,
    },
)
```

### Consuming from a Queue (Go)

```go
msgs, _ := ch.Consume(
    "order.processing",
    "order-service-consumer",
    false, // autoAck — always false; ack manually after processing
    false, false, false, nil,
)

for msg := range msgs {
    if err := processOrder(msg.Body); err != nil {
        msg.Nack(false, false) // send to dead-letter
    } else {
        msg.Ack(false)
    }
}
```

### RPC Pattern (Node.js — currency-service calling pricing-service)

```js
// Caller
const replyQueue = await channel.assertQueue('', { exclusive: true });
const corrId = uuid();
channel.sendToQueue('rpc.pricing.calculate', Buffer.from(JSON.stringify(req)), {
  correlationId: corrId,
  replyTo: replyQueue.queue,
});
// Wait for response on replyQueue filtered by corrId
```

### Delayed Retry (order cancellation reminder after 30 min)

```go
ch.Publish("commerce.orders.delayed", "order.cancel_reminder", false, false,
    amqp.Publishing{
        Headers:      amqp.Table{"x-delay": 1800000}, // 30 minutes in ms
        ContentType:  "application/json",
        DeliveryMode: amqp.Persistent,
        Body:         body,
    })
```

---

## Local Setup

RabbitMQ is included in `docker-compose.yml`. The `definitions.json` is auto-loaded on first
start via `management.load_definitions` in `rabbitmq.conf`.

```bash
# Start RabbitMQ
docker-compose up rabbitmq

# Management UI
open http://localhost:15672   # user: shopos / pass: shopos

# Prometheus metrics (Phase 4 — enable prometheus plugin first)
# open http://localhost:15692/metrics
```

### Required Plugins

```bash
rabbitmq-plugins enable rabbitmq_management
rabbitmq-plugins enable rabbitmq_prometheus
rabbitmq-plugins enable rabbitmq_delayed_message_exchange
```

---

## RabbitMQ vs Kafka in ShopOS

| Concern | RabbitMQ | Kafka |
|---|---|---|
| Primary model | Push (broker delivers to consumer) | Pull (consumer polls broker) |
| Message retention | Until acknowledged (or TTL) | Time/size based log retention |
| Ordering | Per-queue FIFO | Per-partition strict ordering |
| Replay | Not native (use dead-letter) | Native via offset reset |
| Use in ShopOS | Task queues, RPC, fan-out, delayed msgs | Domain events, analytics, CDC |
| Throughput | Millions/day | Millions/second |
| Routing | Powerful (topic, fanout, headers) | Topic + partition key only |
