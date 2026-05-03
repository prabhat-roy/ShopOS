# NATS JetStream — ShopOS Real-Time Messaging

## Role in ShopOS

NATS JetStream handles low-latency pub/sub and real-time event delivery across ShopOS,
filling the gap between Kafka (durable event streaming) and RabbitMQ (task queues). NATS
excels at sub-millisecond message delivery for interactive workloads.

| Domain | NATS Use Case | Stream |
|---|---|---|
| Live Chat | `live-chat-service` WebSocket message relay | `CHAT_MESSAGES` |
| User Presence | Online/offline/typing indicators | `PRESENCE` |
| In-App Notifications | Real-time bell/toast delivery via WebSocket | `NOTIFICATIONS` |
| Page/Click Events | Behavioural signal collection for personalisation | `REALTIME_EVENTS` |
| Inventory Availability | Live stock level changes for storefront | `INVENTORY_UPDATES` |
| Notification Routing | Work-queue to notification-orchestrator | `NOTIFICATIONS` |

---

## Stream Architecture

```mermaid
graph TD
    subgraph Publishers
        LC[live-chat-service\nGo]
        US[user-service\nJava]
        PS[personalization-service\nPython]
        IS[inventory-service\nGo]
        NO[notification-orchestrator\nNode.js]
        ES[event-tracking-service\nPython]
    end

    subgraph JetStream Streams
        CHAT["CHAT_MESSAGES\nchat.> | memory | 100k msgs"]
        PRES["PRESENCE\npresence.> | memory | 10s TTL"]
        RT["REALTIME_EVENTS\nrealtime.> | memory | 1h TTL"]
        NOTIF["NOTIFICATIONS\nnotify.> | file | 24h TTL"]
        INV["INVENTORY_UPDATES\ninventory.> | file | 12h TTL"]
    end

    subgraph Consumers
        LCS[live-chat-service\nWebSocket fanout]
        IAN[in-app-notification-service\nGo]
        NOREC[notification-orchestrator\nroutes to email/sms/push]
        REC[recommendation-service\nPython]
        ET[event-tracking-service\nPython/Cassandra]
    end

    LC -->|"chat.room.{id}.msg"| CHAT --> LCS
    US -->|"presence.user.{id}"| PRES --> LCS
    ES -->|"realtime.product.viewed"| RT --> ET
    PS -->|"realtime.product.clicked"| RT
    NO -->|"notify.inapp.{userId}"| NOTIF --> IAN
    NO -->|"notify.email/sms/push.{userId}"| NOTIF --> NOREC
    IS -->|"inventory.level.{sku}"| INV --> REC
```

---

## Streams Reference

| Stream | Subjects | Storage | Retention | Max Age | Purpose |
|---|---|---|---|---|---|
| `REALTIME_EVENTS` | `realtime.>` | Memory | limits | 1 hour | Page views, clicks, product interactions |
| `NOTIFICATIONS` | `notify.>` | File | workqueue | 24 hours | In-app + channel routing |
| `CHAT_MESSAGES` | `chat.>` | Memory | limits | none (100k msgs) | Live chat relay |
| `PRESENCE` | `presence.>` | Memory | limits | 10 seconds | User online status |
| `INVENTORY_UPDATES` | `inventory.>` | File | limits | 12 hours | Real-time stock levels |

---

## Usage Patterns

### Publish a Chat Message (Go — live-chat-service)

```go
import "github.com/nats-io/nats.go"

nc, _ := nats.Connect("nats://shopos:shopos@nats:4222")
js, _ := nc.JetStream()

msg := ChatMessage{RoomID: "room-42", UserID: "usr-7", Text: "Hello!"}
data, _ := json.Marshal(msg)

// Subject pattern: chat.room.{roomId}.msg
_, err := js.Publish("chat.room.room-42.msg", data)
```

### Subscribe to In-App Notifications (Go — in-app-notification-service)

```go
js.Subscribe("notify.inapp.>",
    func(msg *nats.Msg) {
        // parse UserId from subject: notify.inapp.{userId}
        userId := strings.Split(msg.Subject, ".")[2]
        ws.SendToUser(userId, msg.Data)
        msg.Ack()
    },
    nats.Durable("in-app-notification-consumer"),
    nats.ManualAck(),
    nats.AckWait(10*time.Second),
)
```

### Publish Real-Time Event (Python — event-tracking-service)

```python
import nats, asyncio, json

async def main():
    nc = await nats.connect("nats://shopos:shopos@nats:4222")
    js = nc.jetstream()
    await js.publish("realtime.product.viewed", json.dumps({
        "user_id": "usr-99", "product_id": "prod-512", "ts": 1714000000
    }).encode())
    await nc.close()

asyncio.run(main())
```

### Request-Reply (Node.js — currency-service)

```js
// NATS core (non-JetStream) request-reply — ultra-low latency
const response = await nc.request('rpc.currency.convert',
    JSON.stringify({ amount: 100, from: 'USD', to: 'EUR' }),
    { timeout: 2000 }
);
console.log(JSON.parse(response.data));
```

---

## NATS vs Kafka vs RabbitMQ — Comparison

| Attribute | NATS JetStream | Apache Kafka | RabbitMQ |
|---|---|---|---|
| Primary use | Real-time pub/sub, presence, chat | Event sourcing, CDC, analytics | Task queues, RPC, fan-out |
| Latency | < 1 ms (memory) | 2–10 ms | 1–5 ms |
| Throughput | Very high (millions/sec) | Extremely high | High (millions/day) |
| Message model | At-most / at-least / exactly-once | At-least / exactly-once | At-least / at-most |
| Replay | Yes (JetStream, bounded) | Yes (unlimited log) | Limited (dead-letter) |
| Routing | Subject wildcards (`>`, `*`) | Topic + partition key | Exchanges (topic, fanout, direct, headers) |
| Protocol | NATS wire protocol | Custom TCP / REST | AMQP 0-9-1 |
| Persistence | Optional (memory or file) | Always (log segments) | Optional (durable queues) |
| Cluster model | Raft-based JetStream | ZooKeeper / KRaft | Mirroring / Quorum queues |
| Best for ShopOS | Chat, presence, real-time events | Domain events, CDC, batch analytics | Delayed messages, job queues, RPC |

---

## Local Setup

```bash
# Start NATS JetStream
docker-compose up nats

# NATS CLI — apply stream definitions
nats stream add --config messaging/nats/streams.json

# Monitor via built-in HTTP
open http://localhost:8222

# List streams
nats stream ls

# View stream info
nats stream info CHAT_MESSAGES
```

### Docker Compose Snippet

```yaml
nats:
  image: nats:2.10-alpine
  command: ["--config", "/etc/nats/nats.conf"]
  ports:
    - "4222:4222"   # client port
    - "8222:8222"   # monitoring HTTP
    - "6222:6222"   # cluster routing
  volumes:
    - ./messaging/nats/nats.conf:/etc/nats/nats.conf:ro
    - nats-data:/data
```
