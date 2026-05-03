# Rollback Runbook

> How to undo a deployment safely. Match the rollback method to the deployment method.
> Last reviewed: 2026-05-02.

## Decision tree
```
Was the change ONLY a Helm-chart-version bump?
  ├── yes → Step 1 (Argo Rollouts undo)
  └── no  → Did it include a database migration?
              ├── no  → Step 2 (Argo CD revision rollback)
              └── yes → Step 3 (coordinated rollback with DB)
```

## Step 1 — Argo Rollouts undo (preferred for service-only changes)
```bash
kubectl argo rollouts history rollout/<svc> -n <ns>
kubectl argo rollouts abort   rollout/<svc> -n <ns>
kubectl argo rollouts undo    rollout/<svc> -n <ns>
kubectl argo rollouts get rollout <svc> -n <ns> --watch
```
Time: ~2 minutes. Reversible: yes (just undo again).

## Step 2 — ArgoCD revision rollback (chart values changed)
```bash
argocd app history <svc>-prod
argocd app rollback <svc>-prod <id>
argocd app set <svc>-prod --sync-policy none   # IMPORTANT: pause autosync first
```
Time: ~3 minutes. Reversible: revert by re-enabling auto-sync.

## Step 3 — Coordinated rollback with database migration
**Stop. Page the platform-team-lead before touching anything.** A forward-only migration may have already introduced incompatible schema changes.

Sequence:
1. Pause ArgoCD autosync on the affected app
2. Scale the new version's pods to 0 (`kubectl scale deploy/<svc>-canary --replicas=0`)
3. Decide:
   - Backwards-compatible additive change (new column nullable, new table): roll forward by deploying a fixed app image; do NOT roll back DB
   - Backwards-incompatible change (column dropped, NOT NULL added, type change): write a compensating migration (`flyway migrate -target=<previous>` is rarely safe — usually need a new "undo" migration)
4. Run any compensating migration: `kubectl exec -n flyway flyway-job -- flyway migrate`
5. Resume autosync once schema and app are aligned

## Step 4 — Frontend rollback
```bash
kubectl argo rollouts set image rollout/storefront -n shopos-web storefront=ghcr.io/shopos/storefront:<prev-tag>
scripts/cdn-purge.sh storefront
```

## Step 5 — Infrastructure rollback (Terraform)
```bash
cd infra/terraform/<cloud>/<module>
git checkout HEAD~1 -- .
terraform plan -out=rollback.plan
terraform apply rollback.plan
```
**Never** `terraform destroy` a stateful resource (RDS, S3, Cassandra) as a rollback. Restore from Velero/snapshot instead.

## Verification after rollback
- [ ] ArgoCD app shows Synced/Healthy
- [ ] Pod count back to expected (`kubectl get deploy -n <ns>`)
- [ ] Error rate back to baseline in Grafana for ≥ 10 minutes
- [ ] No new pages from Alertmanager
- [ ] Customer-facing smoke test passes (`scripts/smoke/checkout.sh`)
- [ ] Comms lead posts "Resolved" on Cachet status page

## What NOT to do
- Do not `kubectl delete` a Deployment to "force" rollback — let ArgoCD reconcile
- Do not `git revert` a merged PR before the rollback is verified — leaves chart values out of sync with running pods
- Do not rollback in dev/staging without telling the team — others may be relying on the broken state for debugging
- Do not skip the postmortem because "the rollback worked" — the deploy still broke
