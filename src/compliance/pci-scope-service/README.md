# pci-scope-service

> Tracks PCI-DSS scope: which services touch cardholder data, which env vars are CHD-related.

## Overview

Static + runtime detection (proxy headers); produces PCI scope diagram for auditors

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50285) |
| Health | HTTP `/healthz` |

## Responsibilities

- Tag service in scope
- scope diagram
- evidence collection

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up pci-scope-service
# or
go run .
```

## Health

```bash
curl http://localhost:50285/healthz   # { "status": "ok" }
```

## Related

- Domain: [`compliance`](../)
- Helm chart: [`helm/services/pci-scope-service`](../../../helm/services/pci-scope-service)
- ArgoCD app: `argocd app get pci-scope-service`
