# reorder-service

> One-click reorder of past orders with stock availability and substitution suggestions.

## Overview

Calls inventory-service for stock, suggests alternatives via recommendation-service

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50231) |
| Health | HTTP `/healthz` |

## Responsibilities

- Reorder(order_id)
- list reorderable past orders
- suggest substitutes for OOS items

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up reorder-service
# or
go run .
```

## Health

```bash
curl http://localhost:50231/healthz   # { "status": "ok" }
```

## Related

- Domain: [`commerce`](../)
- Helm chart: [`helm/services/reorder-service`](../../../helm/services/reorder-service)
- ArgoCD app: `argocd app get reorder-service`
