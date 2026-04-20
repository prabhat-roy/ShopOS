# price-alert-service

> Customers subscribe to price-drop alerts on products; the service evaluates subscriptions when pricing-service emits price change events.

## Overview

The price-alert-service lets customers set a target price threshold on any product. Subscriptions are stored in Redis as sorted sets keyed by product, with the customer's target price as the score. When the pricing-service publishes a `catalog.price.updated` Kafka event, this service evaluates all subscriptions for the affected product and fires notification events for customers whose target price has been met or beaten. This architecture decouples price monitoring from checkout and avoids polling the pricing-service.

## Architecture

```mermaid
graph LR
    WB[web-bff :8081] -->|SubscribePriceAlert gRPC| PA[price-alert-service :50186]
    PA -->|ZADD alerts:{product_id} target customer_id| R[(Redis)]
    K[Kafka] -->|catalog.price.updated| PA
    PA -->|ZRANGEBYSCORE &lt;= new_price| R
    PA -->|notification.email.requested| K
    PA -->|notification.push.requested| K
    NO[notification-orchestrator] -->|consume| K
```

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.24 |
| Database | Redis 7 (sorted sets per product) |
| Messaging | Apache Kafka |
| Protocol | gRPC (port 50186) |
| Health Check | HTTP /healthz |

## Key Responsibilities

- Allow customers to create, update, and delete price-drop subscriptions with a target price
- Store subscriptions in Redis sorted sets keyed by product ID with target price as score
- Consume `catalog.price.updated` Kafka events from pricing-service
- Evaluate subscriptions by range query (ZRANGEBYSCORE) against the new price
- Emit `notification.email.requested` and `notification.push.requested` Kafka events for triggered alerts
- Remove triggered subscriptions to prevent repeated notifications for the same price drop
- Enforce a maximum number of active alerts per customer

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GRPC_PORT` | `50186` | gRPC listen port |
| `REDIS_URL` | — | Redis connection URL |
| `KAFKA_BROKERS` | — | Comma-separated Kafka broker addresses |

## Running Locally

```bash
docker-compose up price-alert-service
```

## Health Check

`GET /healthz` → `{"status":"ok"}`

gRPC health: `grpc.health.v1.Health/Check` → `SERVING`
