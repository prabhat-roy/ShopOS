# api-changelog-service

> Tracks API breaking and additive changes from buf-breaking + OpenAPI diffs; publishes changelog.

## Overview

Webhook from CI on every merge to main; Markdown changelog rendered in developer-portal

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | HTTP (port 8225) |
| Health | HTTP `/healthz` |

## Responsibilities

- Receive change event
- render changelog entry
- notify subscribers

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up api-changelog-service
# or
go run .
```

## Health

```bash
curl http://localhost:8225/healthz   # { "status": "ok" }
```

## Related

- Domain: [`developer-platform`](../)
- Helm chart: [`helm/services/api-changelog-service`](../../../helm/services/api-changelog-service)
- ArgoCD app: `argocd app get api-changelog-service`
