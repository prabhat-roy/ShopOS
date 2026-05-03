# account-linking-service

> Merges duplicate user accounts (email, social, passkey) into a single canonical identity.

## Overview

Postgres transactional merge, Kafka outbox events for downstream services

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + Kafka |
| Protocol | gRPC (port 50345) |
| Health | HTTP `/healthz` |

## Responsibilities

- Detect candidate duplicates
- merge with conflict resolution
- emit identity.user.merged event

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up account-linking-service
# or
go run .
```

## Health

```bash
curl http://localhost:50345/healthz   # { "status": "ok" }
```

## Related

- Domain: [`identity`](../)
- Helm chart: [`helm/services/account-linking-service`](../../../helm/services/account-linking-service)
- ArgoCD app: `argocd app get account-linking-service`
