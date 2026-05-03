# passkey-service

> WebAuthn / FIDO2 passkey registration, authentication, and credential lifecycle.

## Overview

go-webauthn library, Postgres for credential storage, Redis for challenge cache

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres + Redis |
| Protocol | gRPC (port 50068) |
| Health | HTTP `/healthz` |

## Responsibilities

- Begin/complete registration
- begin/complete assertion
- list user credentials
- revoke credential

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up passkey-service
# or
go run .
```

## Health

```bash
curl http://localhost:50068/healthz   # { "status": "ok" }
```

## Related

- Domain: [`identity`](../)
- Helm chart: [`helm/services/passkey-service`](../../../helm/services/passkey-service)
- ArgoCD app: `argocd app get passkey-service`
