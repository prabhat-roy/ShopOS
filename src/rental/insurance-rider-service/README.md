# insurance-rider-service

> Optional insurance rider purchase for rentals; integrates with third-party insurer.

## Overview

Quote → bind → claim flow; stores policy doc in MinIO

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + MinIO |
| Protocol | gRPC (port 50324) |
| Health | HTTP `/healthz` |

## Responsibilities

- Quote
- bind
- fileClaim
- policy document

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up insurance-rider-service
# or
go run .
```

## Health

```bash
curl http://localhost:50324/healthz   # { "status": "ok" }
```

## Related

- Domain: [`rental`](../)
- Helm chart: [`helm/services/insurance-rider-service`](../../../helm/services/insurance-rider-service)
- ArgoCD app: `argocd app get insurance-rider-service`
