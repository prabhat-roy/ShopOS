# cash-flow-forecast-service

> Rolling 13-week cash-flow forecast from receivables, payables, payouts, and chargebacks.

## Overview

Time-series model on TimescaleDB; outputs to BI dashboards

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + TimescaleDB |
| Protocol | gRPC (port 50363) |
| Health | HTTP `/healthz` |

## Responsibilities

- Generate forecast
- scenario analysis
- alert on negative weeks

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up cash-flow-forecast-service
# or
go run .
```

## Health

```bash
curl http://localhost:50363/healthz   # { "status": "ok" }
```

## Related

- Domain: [`financial`](../)
- Helm chart: [`helm/services/cash-flow-forecast-service`](../../../helm/services/cash-flow-forecast-service)
- ArgoCD app: `argocd app get cash-flow-forecast-service`
