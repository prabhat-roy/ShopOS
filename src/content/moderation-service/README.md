# moderation-service

> Moderates user-generated content (reviews, Q&A, listings) with rules + ML classifier + human queue.

## Overview

Pipelines through analytics-ai sentiment + custom toxicity classifier; human-review queue in Postgres

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + analytics-ai |
| Protocol | gRPC (port 50149) |
| Health | HTTP `/healthz` |

## Responsibilities

- Classify text/image
- auto-approve / auto-reject
- queue for human review
- appeal flow

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up moderation-service
# or
go run .
```

## Health

```bash
curl http://localhost:50149/healthz   # { "status": "ok" }
```

## Related

- Domain: [`content`](../)
- Helm chart: [`helm/services/moderation-service`](../../../helm/services/moderation-service)
- ArgoCD app: `argocd app get moderation-service`
