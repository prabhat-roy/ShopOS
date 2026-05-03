# cart-recovery-service

> Drives abandoned-cart recovery flow: targeted email/SMS, discount escalation, recovery link.

## Overview

Consumes commerce.cart.abandoned, triggers communications domain

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Redis + Kafka |
| Protocol | gRPC (port 50232) |
| Health | HTTP `/healthz` |

## Responsibilities

- Schedule recovery sequence
- track conversion
- discount escalation ladder

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up cart-recovery-service
# or
go run .
```

## Health

```bash
curl http://localhost:50232/healthz   # { "status": "ok" }
```

## Related

- Domain: [`commerce`](../)
- Helm chart: [`helm/services/cart-recovery-service`](../../../helm/services/cart-recovery-service)
- ArgoCD app: `argocd app get cart-recovery-service`
