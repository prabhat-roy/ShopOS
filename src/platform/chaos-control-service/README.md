# chaos-control-service

> Programmatic API to trigger and schedule Chaos Mesh / Litmus experiments.

## Overview

Chaos Mesh CRD client, Litmus Argo Workflows trigger

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Chaos Mesh CRDs |
| Protocol | gRPC (port 50359) |
| Health | HTTP `/healthz` |

## Responsibilities

- Trigger experiment by name
- schedule game-day
- abort running experiment
- RBAC by domain

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up chaos-control-service
# or
go run .
```

## Health

```bash
curl http://localhost:50359/healthz   # { "status": "ok" }
```

## Related

- Domain: [`platform`](../)
- Helm chart: [`helm/services/chaos-control-service`](../../../helm/services/chaos-control-service)
- ArgoCD app: `argocd app get chaos-control-service`
