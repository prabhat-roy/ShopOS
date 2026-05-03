# tip-service

> Adds gratuity to checkout (delivery, in-store) and tracks per-staff tip distribution.

## Overview

Computes tax-on-tip per jurisdiction, splits across staff

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50230) |
| Health | HTTP `/healthz` |

## Responsibilities

- Calculate tip suggestions
- apply tip to order
- distribute to staff
- report by period

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up tip-service
# or
go run .
```

## Health

```bash
curl http://localhost:50230/healthz   # { "status": "ok" }
```

## Related

- Domain: [`commerce`](../)
- Helm chart: [`helm/services/tip-service`](../../../helm/services/tip-service)
- ArgoCD app: `argocd app get tip-service`
