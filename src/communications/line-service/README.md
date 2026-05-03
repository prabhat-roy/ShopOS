# line-service

> LINE Messaging API integration (Japan/Asia messaging channel).

## Overview

LINE OAuth, message API, rich-content templates

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | LINE Cloud API |
| Protocol | gRPC (port 50133) |
| Health | HTTP `/healthz` |

## Responsibilities

- Send message via LINE
- receive webhook
- rich menu management

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up line-service
# or
go run .
```

## Health

```bash
curl http://localhost:50133/healthz   # { "status": "ok" }
```

## Related

- Domain: [`communications`](../)
- Helm chart: [`helm/services/line-service`](../../../helm/services/line-service)
- ArgoCD app: `argocd app get line-service`
