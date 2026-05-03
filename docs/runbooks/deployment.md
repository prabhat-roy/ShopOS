# Deployment Runbook

> Authoritative procedure for promoting code from `main` → staging → production.
> Last reviewed: 2026-05-02.

## Pipeline overview
1. PR merged to `main` triggers Jenkins pipeline `01-build-test-scan`
2. Image is built with Ko/Kaniko, scanned (Trivy + Grype + Cosign signed), pushed to Harbor + GHCR
3. ArgoCD ApplicationSet detects the new chart version (Renovate bumps `helm/services/<svc>/values.yaml`) and syncs to staging
4. Argo Rollouts performs a canary (10% → 25% → 50% → 100%) with automated SLO analysis
5. Promotion to production is gated by a manual approval in GoCD

## Promote a single service
```bash
# 1. confirm passing in staging
argocd app get <svc>-staging --refresh
argocd app sync <svc>-staging
kubectl rollout status -n <ns> deploy/<svc>

# 2. promote to production via GoCD value-stream view (https://gocd.shopos.example.com)
gocd-cli pipeline trigger promote-prod --pipeline <svc> --counter <build>

# 3. observe rollout
kubectl argo rollouts get rollout <svc> -n <ns> --watch
kubectl argo rollouts dashboard
```

## Promote across all services (release train)
1. Tag main: `git tag -a release-2026.05.0 -m "Release 2026.05.0" && git push --tags`
2. Run `scripts/release-train.sh release-2026.05.0` — bumps all chart versions and opens a single PR
3. After merge, ArgoCD `release-train-prod` Application syncs the bundle in dependency order (platform → identity → catalog → commerce → ...)
4. Pyrra dashboards must show error budget burn < 1× for 30m before declaring success

## Pre-production checks
- [ ] CI green on `main` (Jenkins + GitHub Actions PR-validation)
- [ ] No active P1/P2 incidents (PagerDuty)
- [ ] Backstage Tech Insights score green for service
- [ ] No Renovate-pinned security advisories pending
- [ ] No active feature flag rollouts blocking the change (Unleash UI)
- [ ] Pact contract verification passes against the latest provider tag
- [ ] OpenAPI/Buf breaking-change check passes

## Rollout cadence
| Environment | Cadence       | Approver        |
|-------------|--------------|-----------------|
| dev         | on every push | none            |
| staging     | on merge to main | none          |
| production  | manual gate (GoCD) | release manager |

## Failure handling
- If Argo Rollouts auto-pauses (SLO regression): see `rollback.md` step 1
- If db migration step fails: stop rollout, run `flyway info` to see state, escalate to platform on-call
- If post-deploy alerts fire: follow `incident-response.md`

## Common commands
```bash
# Inspect ApplicationSet status
argocd appset get all-services
# Diff what would be deployed
argocd app diff <svc>-prod
# Force re-sync a stuck app
argocd app sync <svc>-prod --force --replace
# Pause auto-sync during incident
argocd app set <svc>-prod --sync-policy none
```

## Out-of-hours releases
Discouraged. If unavoidable:
1. Open #release-bridge in Slack, page release-manager-oncall
2. Skip pre-production smoke tests requires platform-team-lead approval (recorded as Linear comment)
3. Hold production for 30 minutes post-rollout to monitor — do not log off
