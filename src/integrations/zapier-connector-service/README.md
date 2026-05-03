# zapier-connector-service

> Public Zapier app — exposes ShopOS triggers (new order, low stock) and actions (create product).

## Overview

OAuth via developer-platform, webhook subscriptions, polling endpoints

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | developer-platform |
| Protocol | HTTP (port 8222) |
| Health | HTTP `/healthz` |

## Responsibilities

- List triggers
- list actions
- subscribe to webhook
- execute action

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up zapier-connector-service
# or
go run .
```

## Health

```bash
curl http://localhost:8222/healthz   # { "status": "ok" }
```

## Related

- Domain: [`integrations`](../)
- Helm chart: [`helm/services/zapier-connector-service`](../../../helm/services/zapier-connector-service)
- ArgoCD app: `argocd app get zapier-connector-service`
