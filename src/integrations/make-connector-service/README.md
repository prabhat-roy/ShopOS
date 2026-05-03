# make-connector-service

> Make.com (Integromat) integration analogous to Zapier connector.

## Overview

Same OAuth + webhook + action model; tailored to Make's IML format

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | developer-platform |
| Protocol | HTTP (port 8223) |
| Health | HTTP `/healthz` |

## Responsibilities

- List modules
- subscribe to instant trigger
- polling trigger

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up make-connector-service
# or
go run .
```

## Health

```bash
curl http://localhost:8223/healthz   # { "status": "ok" }
```

## Related

- Domain: [`integrations`](../)
- Helm chart: [`helm/services/make-connector-service`](../../../helm/services/make-connector-service)
- ArgoCD app: `argocd app get make-connector-service`
