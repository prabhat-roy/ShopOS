# waitlist-event-service

> Waitlist for sold-out events; notifies + auto-purchases when seats become available.

## Overview

Redis sorted set per event; consumes ticket release events

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Redis |
| Protocol | gRPC (port 50306) |
| Health | HTTP `/healthz` |

## Responsibilities

- Join waitlist
- leave waitlist
- auto-buy on release
- expire window

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up waitlist-event-service
# or
go run .
```

## Health

```bash
curl http://localhost:50306/healthz   # { "status": "ok" }
```

## Related

- Domain: [`events-ticketing`](../)
- Helm chart: [`helm/services/waitlist-event-service`](../../../helm/services/waitlist-event-service)
- ArgoCD app: `argocd app get waitlist-event-service`
