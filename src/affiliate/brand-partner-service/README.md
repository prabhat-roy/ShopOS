# brand-partner-service

> Brand-tier partnerships: co-marketing campaigns, exclusive products, branded landing pages.

## Overview

Postgres for partnership state, integrates with affiliate-service for tracking

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50249) |
| Health | HTTP `/healthz` |

## Responsibilities

- RegisterPartner
- createCampaign
- trackPartnerSales

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up brand-partner-service
# or
go run .
```

## Health

```bash
curl http://localhost:50249/healthz   # { "status": "ok" }
```

## Related

- Domain: [`affiliate`](../)
- Helm chart: [`helm/services/brand-partner-service`](../../../helm/services/brand-partner-service)
- ArgoCD app: `argocd app get brand-partner-service`
