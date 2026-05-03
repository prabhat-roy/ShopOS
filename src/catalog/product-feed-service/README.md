# product-feed-service

> Generates Google Shopping, Meta, and TikTok product feeds on a schedule.

## Overview

S3/MinIO upload, XML/CSV/JSON formats, incremental delta feeds

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + MinIO |
| Protocol | HTTP (port 8221) |
| Health | HTTP `/healthz` |

## Responsibilities

- Generate full feed (daily)
- delta feed (15min)
- validate against destination schema
- expose feed URL

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up product-feed-service
# or
go run .
```

## Health

```bash
curl http://localhost:8221/healthz   # { "status": "ok" }
```

## Related

- Domain: [`catalog`](../)
- Helm chart: [`helm/services/product-feed-service`](../../../helm/services/product-feed-service)
- ArgoCD app: `argocd app get product-feed-service`
