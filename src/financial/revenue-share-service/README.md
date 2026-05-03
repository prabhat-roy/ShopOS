# revenue-share-service

> Splits inbound revenue across partners (marketplace seller, affiliate, platform fee, tax).

## Overview

Rule engine with per-contract overrides, idempotent ledger writes

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50364) |
| Health | HTTP `/healthz` |

## Responsibilities

- ComputeSplit(order)
- preview split
- settle period
- audit ledger

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up revenue-share-service
# or
go run .
```

## Health

```bash
curl http://localhost:50364/healthz   # { "status": "ok" }
```

## Related

- Domain: [`financial`](../)
- Helm chart: [`helm/services/revenue-share-service`](../../../helm/services/revenue-share-service)
- ArgoCD app: `argocd app get revenue-share-service`
