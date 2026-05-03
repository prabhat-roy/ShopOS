# soc2-evidence-service

> Continuously collects SOC2 evidence (access logs, change tickets, scan reports) into a single audit pack.

## Overview

Cron-based collectors + on-demand audit-pack export

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + MinIO |
| Protocol | gRPC (port 50286) |
| Health | HTTP `/healthz` |

## Responsibilities

- Collect evidence
- generate audit pack
- auditor share-link

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up soc2-evidence-service
# or
go run .
```

## Health

```bash
curl http://localhost:50286/healthz   # { "status": "ok" }
```

## Related

- Domain: [`compliance`](../)
- Helm chart: [`helm/services/soc2-evidence-service`](../../../helm/services/soc2-evidence-service)
- ArgoCD app: `argocd app get soc2-evidence-service`
