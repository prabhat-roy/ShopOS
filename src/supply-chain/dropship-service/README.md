# dropship-service

> Vendor-direct fulfillment: passes orders to vendor APIs, tracks status, handles split shipments.

## Overview

EDI / REST adapters per vendor, status webhook receiver

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + Kafka |
| Protocol | gRPC (port 50379) |
| Health | HTTP `/healthz` |

## Responsibilities

- RouteOrder
- trackShipment
- handleSplitShipment
- vendor SLA monitoring

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up dropship-service
# or
go run .
```

## Health

```bash
curl http://localhost:50379/healthz   # { "status": "ok" }
```

## Related

- Domain: [`supply-chain`](../)
- Helm chart: [`helm/services/dropship-service`](../../../helm/services/dropship-service)
- ArgoCD app: `argocd app get dropship-service`
