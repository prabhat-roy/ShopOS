# service-registry-service

> Thin wrapper around Consul for service discovery and health-check aggregation.

## Overview

Consul SDK, gRPC façade for non-Consul-aware clients

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Consul |
| Protocol | gRPC (port 50357) |
| Health | HTTP `/healthz` |

## Responsibilities

- List services by domain
- register/deregister
- health-check propagation
- tag-based filtering

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up service-registry-service
# or
go run .
```

## Health

```bash
curl http://localhost:50357/healthz   # { "status": "ok" }
```

## Related

- Domain: [`platform`](../)
- Helm chart: [`helm/services/service-registry-service`](../../../helm/services/service-registry-service)
- ArgoCD app: `argocd app get service-registry-service`
