# seller-tier-service

> Maintains seller tiers (bronze/silver/gold) based on volume, ratings, and disputes.

## Overview

Daily batch recompute; tier drives commission rate via marketplace-commission-service

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50257) |
| Health | HTTP `/healthz` |

## Responsibilities

- GetTier(seller)
- recompute (cron)
- manual tier override (admin)

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up seller-tier-service
# or
go run .
```

## Health

```bash
curl http://localhost:50257/healthz   # { "status": "ok" }
```

## Related

- Domain: [`marketplace`](../)
- Helm chart: [`helm/services/seller-tier-service`](../../../helm/services/seller-tier-service)
- ArgoCD app: `argocd app get seller-tier-service`
