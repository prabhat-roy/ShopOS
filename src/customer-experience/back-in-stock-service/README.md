# back-in-stock-service

> Customers subscribe to restock notifications; the service fans out alerts to all subscribers when inventory-service emits a restock event.

## Overview

The back-in-stock-service maintains a per-product subscriber list in Redis. When a customer opts in for a "notify me when available" alert, their customer ID is appended to the product's subscriber set. The service consumes `supplychain.inventory.restocked` Kafka events and immediately publishes notification events for every subscriber, enabling the notification-orchestrator to dispatch emails and push notifications. Unlike waitlist-service (which reserves stock for a single next customer), this service broadcasts to all subscribers without reserving inventory.

## Architecture

```mermaid
graph LR
    WB[web-bff :8081] -->|SubscribeRestock gRPC| BS[back-in-stock-service :50187]
    BS -->|SADD subscribers:{product_id} customer_id| R[(Redis)]
    K[Kafka] -->|supplychain.inventory.restocked| BS
    BS -->|SMEMBERS subscribers:{product_id}| R
    BS -->|notification.email.requested x N| K
    BS -->|notification.push.requested x N| K
    BS -->|DEL subscribers:{product_id}| R
    NO[notification-orchestrator] -->|consume| K
```

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.24 |
| Database | Redis 7 (sets per product) |
| Messaging | Apache Kafka |
| Protocol | gRPC (port 50187) |
| Health Check | HTTP /healthz |

## Key Responsibilities

- Register and deregister customer subscriptions for out-of-stock products
- Store subscriber sets in Redis with per-product keys for O(1) membership checks
- Consume `supplychain.inventory.restocked` Kafka events
- Fan out `notification.email.requested` and `notification.push.requested` events for all subscribers
- Clear the subscriber set after fan-out to avoid duplicate alerts on future restock events
- Allow customers to check or cancel their existing subscriptions via gRPC
- Enforce a per-customer subscription limit to prevent abuse

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GRPC_PORT` | `50187` | gRPC listen port |
| `REDIS_URL` | — | Redis connection URL |
| `KAFKA_BROKERS` | — | Comma-separated Kafka broker addresses |

## Running Locally

```bash
docker-compose up back-in-stock-service
```

## Health Check

`GET /healthz` → `{"status":"ok"}`

gRPC health: `grpc.health.v1.Health/Check` → `SERVING`
