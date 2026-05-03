# catalog-translation-service

> Auto-translates product titles and descriptions across locales using a translation provider.

## Overview

DeepL/AWS Translate adapter, MongoDB cache, cost-tracking

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | MongoDB + provider |
| Protocol | gRPC (port 50376) |
| Health | HTTP `/healthz` |

## Responsibilities

- Translate(sku, target_lang)
- bulk translate by category
- human-review queue
- cost report per locale

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up catalog-translation-service
# or
go run .
```

## Health

```bash
curl http://localhost:50376/healthz   # { "status": "ok" }
```

## Related

- Domain: [`catalog`](../)
- Helm chart: [`helm/services/catalog-translation-service`](../../../helm/services/catalog-translation-service)
- ArgoCD app: `argocd app get catalog-translation-service`
