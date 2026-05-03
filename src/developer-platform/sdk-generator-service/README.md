# sdk-generator-service

> Generates language SDKs (TypeScript, Python, Go, Java) from OpenAPI + buf descriptors.

## Overview

openapi-generator + buf code-gen, publishes to npm/PyPI/Maven/Go module proxy

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | OpenAPI + Buf |
| Protocol | HTTP (port 8224) |
| Health | HTTP `/healthz` |

## Responsibilities

- Generate SDK for version
- publish to registry
- diff vs prev version

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up sdk-generator-service
# or
go run .
```

## Health

```bash
curl http://localhost:8224/healthz   # { "status": "ok" }
```

## Related

- Domain: [`developer-platform`](../)
- Helm chart: [`helm/services/sdk-generator-service`](../../../helm/services/sdk-generator-service)
- ArgoCD app: `argocd app get sdk-generator-service`
