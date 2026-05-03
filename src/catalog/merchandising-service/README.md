# merchandising-service

> Manages merchandising rules: hero banners, featured collections, sort overrides per category.

## Overview

Rule DSL stored in Postgres, served from Redis cache to storefront

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + Redis |
| Protocol | gRPC (port 50377) |
| Health | HTTP `/healthz` |

## Responsibilities

- CRUD rules
- evaluate rules per page
- A/B test merchandising sets

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up merchandising-service
# or
go run .
```

## Health

```bash
curl http://localhost:50377/healthz   # { "status": "ok" }
```

## Related

- Domain: [`catalog`](../)
- Helm chart: [`helm/services/merchandising-service`](../../../helm/services/merchandising-service)
- ArgoCD app: `argocd app get merchandising-service`
