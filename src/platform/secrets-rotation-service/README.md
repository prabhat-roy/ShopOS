# secrets-rotation-service

> Rotates Vault dynamic credentials, JWT signing keys, and DB passwords on a schedule.

## Overview

Vault dynamic engines, K8s CronJob, Postgres, Redis cache

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + Vault |
| Protocol | gRPC (port 50356) |
| Health | HTTP `/healthz` |

## Responsibilities

- Vault auth via K8s ServiceAccount
- Postgres password rotation
- JWT JWKS rotation
- audit log every rotation

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up secrets-rotation-service
# or
go run .
```

## Health

```bash
curl http://localhost:50356/healthz   # { "status": "ok" }
```

## Related

- Domain: [`platform`](../)
- Helm chart: [`helm/services/secrets-rotation-service`](../../../helm/services/secrets-rotation-service)
- ArgoCD app: `argocd app get secrets-rotation-service`
