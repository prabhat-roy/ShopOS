# returns-grading-service

> Grades returned items A/B/C/scrap and routes to restock, refurbish, liquidation, or disposal.

## Overview

Rule engine + image classification (via analytics-ai), updates inventory routing

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50378) |
| Health | HTTP `/healthz` |

## Responsibilities

- Grade(item, photos)
- route to disposition
- track per-vendor return reasons

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up returns-grading-service
# or
go run .
```

## Health

```bash
curl http://localhost:50378/healthz   # { "status": "ok" }
```

## Related

- Domain: [`supply-chain`](../)
- Helm chart: [`helm/services/returns-grading-service`](../../../helm/services/returns-grading-service)
- ArgoCD app: `argocd app get returns-grading-service`
