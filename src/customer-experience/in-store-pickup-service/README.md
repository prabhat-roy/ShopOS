# in-store-pickup-service

> Buy Online Pickup In Store (BOPIS) flow: store selection, pickup window, ready-to-collect notifications.

## Overview

Calls warehouse-service for store inventory, communications for ready alerts

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50237) |
| Health | HTTP `/healthz` |

## Responsibilities

- List nearby stores with stock
- reserve item at store
- mark ready for pickup
- complete pickup

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up in-store-pickup-service
# or
go run .
```

## Health

```bash
curl http://localhost:50237/healthz   # { "status": "ok" }
```

## Related

- Domain: [`customer-experience`](../)
- Helm chart: [`helm/services/in-store-pickup-service`](../../../helm/services/in-store-pickup-service)
- ArgoCD app: `argocd app get in-store-pickup-service`
