# risk-scoring-service

> Computes a 0-100 login risk score from device fingerprint, geo, velocity, and behaviour signals.

## Overview

Online ML inference, Redis sliding-window counters, feature store integration

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Redis + ml-feature-store |
| Protocol | gRPC (port 50069) |
| Health | HTTP `/healthz` |

## Responsibilities

- Score(login_event)
- feedback loop on confirmed fraud
- rule overrides per tenant

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up risk-scoring-service
# or
go run .
```

## Health

```bash
curl http://localhost:50069/healthz   # { "status": "ok" }
```

## Related

- Domain: [`identity`](../)
- Helm chart: [`helm/services/risk-scoring-service`](../../../helm/services/risk-scoring-service)
- ArgoCD app: `argocd app get risk-scoring-service`
