# distributed-lock-service

> Redis-backed distributed lock primitive for cross-pod critical sections.

## Overview

Redis Redlock, gRPC, OpenTelemetry tracing on every Acquire/Release

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Redis |
| Protocol | gRPC (port 50358) |
| Health | HTTP `/healthz` |

## Responsibilities

- Acquire(key, ttl)
- Release(key, fence)
- WaitForLock
- auto-renew lease

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up distributed-lock-service
# or
go run .
```

## Health

```bash
curl http://localhost:50358/healthz   # { "status": "ok" }
```

## Related

- Domain: [`platform`](../)
- Helm chart: [`helm/services/distributed-lock-service`](../../../helm/services/distributed-lock-service)
- ArgoCD app: `argocd app get distributed-lock-service`
