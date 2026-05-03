# seller-onboarding-service

> Walks new sellers through KYC, store setup, payout configuration, and first listing.

## Overview

State-machine workflow stored in Postgres; integrates with kyc-aml + payout services

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + kyc-aml |
| Protocol | gRPC (port 50256) |
| Health | HTTP `/healthz` |

## Responsibilities

- Start onboarding
- submit KYC
- configure payout
- approve seller

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up seller-onboarding-service
# or
go run .
```

## Health

```bash
curl http://localhost:50256/healthz   # { "status": "ok" }
```

## Related

- Domain: [`marketplace`](../)
- Helm chart: [`helm/services/seller-onboarding-service`](../../../helm/services/seller-onboarding-service)
- ArgoCD app: `argocd app get seller-onboarding-service`
