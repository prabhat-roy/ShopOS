# graphql-federation-service

> Apollo Federation gateway composing schemas from per-domain GraphQL services.

## Overview

GraphQL schema stitching, query planning, response caching

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | subgraph services |
| Protocol | HTTP (port 8220) |
| Health | HTTP `/healthz` |

## Responsibilities

- Compose per-domain subgraphs
- query plan
- response cache
- telemetry per resolver

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up graphql-federation-service
# or
go run .
```

## Health

```bash
curl http://localhost:8220/healthz   # { "status": "ok" }
```

## Related

- Domain: [`platform`](../)
- Helm chart: [`helm/services/graphql-federation-service`](../../../helm/services/graphql-federation-service)
- ArgoCD app: `argocd app get graphql-federation-service`
