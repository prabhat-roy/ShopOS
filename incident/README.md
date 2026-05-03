# Incident Management — ShopOS

End-to-end incident workflow: on-call paging → war-room channel → public status page.

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [grafana-oncall/](grafana-oncall/) | Grafana OnCall | On-call rotations, escalation policies, Slack alerts, PagerDuty sync |
| [grafana-incident/](grafana-incident/) | Grafana Incident | War-room creation, role assignment (IC, Comms, Scribe), timeline capture |
| [cachet/](cachet/) | Cachet | Public-facing status page; updated by Uptime Kuma + Grafana Incident hooks |

## Workflow

```
Alertmanager fires → Grafana OnCall pages primary → Slack #incident-* channel created
                                       │
                                       ├─► Grafana Incident assigns IC / Comms / Scribe roles
                                       ├─► Cachet status page auto-updates (degraded → outage)
                                       └─► Postmortem drafted to docs/postmortems/YYYY-MM-DD-*.md
```

## Runbooks

- [Incident response](../docs/runbooks/incident-response.md)
- [Rollback](../docs/runbooks/rollback.md)
- [Postgres failover](../docs/runbooks/postgres-failover.md)
- [Kafka consumer lag](../docs/runbooks/kafka-consumer-lag.md)
