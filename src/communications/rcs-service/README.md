# rcs-service

> Rich Communication Services (RCS) outbound — replaces SMS on Android with rich content.

## Overview

Google RCS Business Messaging API; falls back to sms-service on unsupported devices

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | RCS provider |
| Protocol | gRPC (port 50134) |
| Health | HTTP `/healthz` |

## Responsibilities

- SendRCS
- fallback to SMS
- delivery receipts
- rich card templates

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up rcs-service
# or
go run .
```

## Health

```bash
curl http://localhost:50134/healthz   # { "status": "ok" }
```

## Related

- Domain: [`communications`](../)
- Helm chart: [`helm/services/rcs-service`](../../../helm/services/rcs-service)
- ArgoCD app: `argocd app get rcs-service`
