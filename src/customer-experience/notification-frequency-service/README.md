# notification-frequency-service

> Caps notification frequency per user / per channel to prevent fatigue.

## Overview

Redis sliding-window counters; consulted by notification-orchestrator before send

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Redis |
| Protocol | gRPC (port 50238) |
| Health | HTTP `/healthz` |

## Responsibilities

- CanSend(user, channel)
- record sent
- per-tenant policy override

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up notification-frequency-service
# or
go run .
```

## Health

```bash
curl http://localhost:50238/healthz   # { "status": "ok" }
```

## Related

- Domain: [`customer-experience`](../)
- Helm chart: [`helm/services/notification-frequency-service`](../../../helm/services/notification-frequency-service)
- ArgoCD app: `argocd app get notification-frequency-service`
