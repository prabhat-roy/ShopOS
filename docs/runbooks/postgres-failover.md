# Postgres Patroni Failover Runbook

> When the primary is misbehaving (replication broken, OOM, disk full, hung).
> Last reviewed: 2026-05-02.

## Pre-flight
- Verify the issue is the primary and not the application: `psql -h postgres-primary -c "SELECT 1"`
- Check Patroni cluster health: `patronictl -c /etc/patroni/config.yml list`
- Confirm at least one synchronous replica is `streaming` and lag is < 1MB

## Planned switchover (zero data loss)
```bash
# from any patroni node
patronictl -c /etc/patroni/config.yml switchover
# choose target: pick the synchronous replica with lowest lag
# wait for completion, then verify
patronictl list
```
PgBouncer (`pgbouncer.databases.svc:6432`) will reconnect automatically.

## Emergency failover (potential data loss)
Only if the primary is hard-down:
```bash
patronictl -c /etc/patroni/config.yml failover
# select replica with lowest lag
```
Then:
1. Investigate what data was lost (check WAL archive in MinIO `wal/`)
2. Update Cachet status with confirmed data loss window
3. Trigger reconciliation against the event store (`kubectl create job reconcile-orders --from=cronjob/reconciliation`)

## Post-failover
- [ ] Update DNS or service discovery if PgBouncer doesn't auto-reconnect
- [ ] Re-add the old primary as a replica: `patronictl reinit <oldprimary>`
- [ ] Confirm Velero backup ran on the new primary
- [ ] Check replica lag returns < 100ms within 5 minutes
- [ ] File an incident postmortem (Sev2 minimum)
