# quest-service

> Multi-step quests (e.g. 'first 3 purchases get 2x points') with progress tracking and rewards.

## Overview

State machine per (user, quest); emits gamification.quest.completed event

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.23 |
| Framework | net/http |
| Storage | Postgres |
| Protocol | gRPC (port 50266) |
| Health | HTTP `/healthz` |

## Responsibilities

- StartQuest
- RecordEvent
- CompleteQuest
- list active quests

## Environment Variables

See [.env.example](.env.example).

## Running Locally

```bash
docker-compose up quest-service
# or
go run .
```

## Health

```bash
curl http://localhost:50266/healthz   # { "status": "ok" }
```

## Related

- Domain: [`gamification`](../)
- Helm chart: [`helm/services/quest-service`](../../../helm/services/quest-service)
- ArgoCD app: `argocd app get quest-service`
