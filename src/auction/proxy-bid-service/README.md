# proxy-bid-service

> Manages proxy bids (bid up to a max amount automatically) for auctions.

## Overview

Listens for new bids, places counter-bids on behalf of holder, idempotent

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Redis |
| Protocol | gRPC (port 50314) |
| Health | HTTP `/healthz` |

## Responsibilities

- Set proxy
- cancel proxy
- execute counter-bid

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up proxy-bid-service
# or
go run .
```

## Health

```bash
curl http://localhost:50314/healthz   # { "status": "ok" }
```

## Related

- Domain: [`auction`](../)
- Helm chart: [`helm/services/proxy-bid-service`](../../../helm/services/proxy-bid-service)
- ArgoCD app: `argocd app get proxy-bid-service`
