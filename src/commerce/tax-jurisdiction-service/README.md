# tax-jurisdiction-service

> Resolves tax jurisdiction (rooftop, ZIP+4, postal-zone) for any address worldwide.

## Overview

Local jurisdiction tables, fallback to provider (Avalara/TaxJar) for edge cases

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50233) |
| Health | HTTP `/healthz` |

## Responsibilities

- Resolve(address)
- bulk resolve
- explain rule applied
- audit trail

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up tax-jurisdiction-service
# or
go run .
```

## Health

```bash
curl http://localhost:50233/healthz   # { "status": "ok" }
```

## Related

- Domain: [`commerce`](../)
- Helm chart: [`helm/services/tax-jurisdiction-service`](../../../helm/services/tax-jurisdiction-service)
- ArgoCD app: `argocd app get tax-jurisdiction-service`
