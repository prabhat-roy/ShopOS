# dam-service

> Digital Asset Management — taxonomy, tagging, search, and rights-management for media assets.

## Overview

MinIO storage, Elasticsearch index, image-processing for thumbnails/variants

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | MinIO + Elasticsearch |
| Protocol | gRPC (port 50150) |
| Health | HTTP `/healthz` |

## Responsibilities

- Upload asset with metadata
- search by tags
- rights/expiry tracking
- variant generation

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up dam-service
# or
go run .
```

## Health

```bash
curl http://localhost:50150/healthz   # { "status": "ok" }
```

## Related

- Domain: [`content`](../)
- Helm chart: [`helm/services/dam-service`](../../../helm/services/dam-service)
- ArgoCD app: `argocd app get dam-service`
