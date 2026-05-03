# Incident Response Runbook

> Single source of truth during an incident. Follow this even if it feels excessive.
> Last reviewed: 2026-05-02.

## Severity definitions
| Sev  | Definition                                          | First response  |
|------|-----------------------------------------------------|-----------------|
| Sev1 | Customer-facing outage; revenue impact > $1k/min   | Page primary on-call immediately; war room within 5 min |
| Sev2 | Major feature broken; degraded checkout/auth/search | Page primary on-call within 5 min; Slack #incident |
| Sev3 | Single-domain regression; non-customer-impacting   | Slack #ops; ticket within 30 min |
| Sev4 | Cosmetic / future risk                              | Linear ticket; no page |

## First 5 minutes
1. Page primary on-call (Grafana OnCall escalation policy `prod-primary`)
2. Open #incident-YYYYMMDD-NNN in Slack (run `/incident-bot start sev2 cart 5xx spike`)
3. Assign the three roles in Slack:
   - Incident Commander (IC) — drives the response
   - Communications Lead — updates Cachet status page + customer comms
   - Scribe — pinned thread, captures every action with timestamps
4. Snapshot dashboards before you investigate (Grafana `Share → Snapshot`)

## Investigation playbook
| Symptom                          | First thing to check                  |
|----------------------------------|---------------------------------------|
| 5xx spike across many services   | Istio mesh, then Envoy upstream cluster, then DB |
| Single-service 5xx spike         | Pod count + recent deploy + Sentry top error |
| Slow API responses               | Pyroscope CPU/heap, Tempo top spans, DB slow query log |
| Search returning 0 results       | Elasticsearch cluster health (red?), Meilisearch indexer lag |
| Payments failing                 | Stripe/PayPal status page, payment-service Vault lease, DB conn pool |
| Cart 503s                        | Redis memory + Dragonfly latency + cart-service replicas |
| Notifications not delivered      | Kafka consumer lag (`notification.email.requested`), provider rate limit |

## Common runbook entries
- `https://runbooks.shopos.example.com/identity/auth-5xx`
- `https://runbooks.shopos.example.com/commerce/payment-failures`
- `https://runbooks.shopos.example.com/catalog/search-degraded`
- `https://runbooks.shopos.example.com/infra/postgres-failover`

## During the incident
- Every action goes into the Slack thread BEFORE doing it (`I'm about to scale cart-service to 12 replicas`)
- Hold a checkpoint every 15 minutes — IC restates current hypothesis + next 3 actions
- Communications lead posts a customer-facing update on Cachet every 30 minutes minimum
- If you're stuck for >20 minutes, escalate (don't wait until you've tried everything)

## Mitigation patterns
| Action                        | Tool                       | Reversibility  |
|-------------------------------|----------------------------|----------------|
| Roll back the most recent deploy | argo rollouts undo      | Trivially reversible |
| Scale up replicas              | `kubectl scale`           | Trivially reversible |
| Disable a feature flag         | Unleash UI                | Reversible     |
| Block traffic via Istio        | VirtualService weight=0   | Reversible     |
| Drain a node                   | `kubectl drain`           | Reversible     |
| Failover Postgres              | patronictl switchover     | Reversible (with care) |
| Roll back DB migration         | flyway undo               | Sometimes destructive — pause first |

## Resolving
1. Confirm metrics back to baseline for ≥ 15 minutes
2. Confirm no related alerts fired in the last 15 minutes
3. Post resolution update to Cachet status page
4. Run `/incident-bot resolve` in Slack
5. Schedule postmortem within 5 business days (Sev1) / 10 business days (Sev2)

## Postmortem expectations
- Blameless. Focus on system gaps, not individuals.
- Includes timeline, contributing factors, what worked, what didn't, action items with owners + due dates
- Action items get tracked in Linear with the `incident-followup` label and reviewed weekly until closed
- Postmortem doc lives at `docs/postmortems/YYYY-MM-DD-<slug>.md` and is read by the whole platform team
