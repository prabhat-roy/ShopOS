# circular-economy-service

> Tracks repair, resale, refurbish, recycle flows; reports diverted-from-landfill metric.

## Overview

Integrates with returns-grading + warehouse for disposition data

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50295) |
| Health | HTTP `/healthz` |

## Responsibilities

- RecordDisposition
- compute diversion rate
- report by SKU/category

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up circular-economy-service
# or
go run .
```

## Health

```bash
curl http://localhost:50295/healthz   # { "status": "ok" }
```

## Related

- Domain: [`sustainability`](../)
- Helm chart: [`helm/services/circular-economy-service`](../../../helm/services/circular-economy-service)
- ArgoCD app: `argocd app get circular-economy-service`
