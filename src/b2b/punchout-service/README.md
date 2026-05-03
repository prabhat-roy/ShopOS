# punchout-service

> OCI / cXML PunchOut integration — lets B2B buyers shop ShopOS from inside their procurement system.

## Overview

OCI 4/5 + cXML PunchOutSetup/OrderRequest; session per buyer

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | HTTP (port 8226) |
| Health | HTTP `/healthz` |

## Responsibilities

- StartPunchOut session
- return cart to procurement
- purchase order ingestion

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up punchout-service
# or
go run .
```

## Health

```bash
curl http://localhost:8226/healthz   # { "status": "ok" }
```

## Related

- Domain: [`b2b`](../)
- Helm chart: [`helm/services/punchout-service`](../../../helm/services/punchout-service)
- ArgoCD app: `argocd app get punchout-service`
