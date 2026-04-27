# GitOps â€” ShopOS

ShopOS follows the GitOps operating model: Git is the single source of truth for both
application config and infrastructure state. All deployments are driven by reconciliation
loops rather than imperative scripts. The GitOps toolchain is built entirely on the Argo
project and Flux CD.

---

## Directory Structure

```
gitops/
â”œâ”€â”€ charts/                             â† Standalone Helm charts to install GitOps tools
â”‚   â”œâ”€â”€ argocd/                         â† ArgoCD chart (quay.io/argoproj/argocd:v2.12.0)
â”‚   â”œâ”€â”€ argo-rollouts/                  â† Argo Rollouts chart
â”‚   â”œâ”€â”€ argo-workflows/                 â† Argo Workflows chart
â”‚   â”œâ”€â”€ argo-events/                    â† Argo Events chart
â”‚   â”œâ”€â”€ argocd-image-updater/           â† ArgoCD Image Updater chart
â”‚   â”œâ”€â”€ fluxcd/                         â† Flux CD chart
â”‚   â”œâ”€â”€ flagger/                        â† Flagger progressive delivery chart
â”‚   â”œâ”€â”€ weave-gitops/                   â† Weave GitOps dashboard chart
â”‚   â”œâ”€â”€ sealed-secrets/                 â† Sealed Secrets controller chart
â”‚   â”œâ”€â”€ external-secrets/               â† External Secrets Operator chart
â”‚   â”œâ”€â”€ vcluster/                       â† vCluster virtual K8s chart
â”‚   â””â”€â”€ gimlet/                         â† Gimlet developer platform chart
â”‚
â”œâ”€â”€ argocd/
â”‚   â”œâ”€â”€ app-of-apps.yaml                â† Root ArgoCD Application (bootstraps everything)
â”‚   â”œâ”€â”€ applicationsets/
â”‚   â”‚   â””â”€â”€ all-services.yaml           â† ApplicationSet covering all 154 services
â”‚   â””â”€â”€ projects/                       â† AppProject per domain (13 total)
â”‚       â”œâ”€â”€ platform-project.yaml
â”‚       â”œâ”€â”€ identity-project.yaml
â”‚       â”œâ”€â”€ catalog-project.yaml
â”‚       â”œâ”€â”€ commerce-project.yaml
â”‚       â”œâ”€â”€ supply-chain-project.yaml
â”‚       â”œâ”€â”€ financial-project.yaml
â”‚       â”œâ”€â”€ customer-experience-project.yaml
â”‚       â”œâ”€â”€ communications-project.yaml
â”‚       â”œâ”€â”€ content-project.yaml
â”‚       â”œâ”€â”€ analytics-ai-project.yaml
â”‚       â”œâ”€â”€ b2b-project.yaml
â”‚       â”œâ”€â”€ integrations-project.yaml
â”‚       â””â”€â”€ affiliate-project.yaml
â”‚
â”œâ”€â”€ flux/
â”‚   â”œâ”€â”€ base/                           â† Shared Flux resources (GitRepository + HelmReleases)
â”‚   â”‚   â”œâ”€â”€ kustomization.yaml
â”‚   â”‚   â”œâ”€â”€ gitrepository.yaml
â”‚   â”‚   â””â”€â”€ helm-releases.yaml          â† HelmRelease objects for all key services
â”‚   â””â”€â”€ clusters/
â”‚       â”œâ”€â”€ production/                 â† Production overlay (replica=3, higher resources)
â”‚       â”‚   â”œâ”€â”€ kustomization.yaml
â”‚       â”‚   â””â”€â”€ namespaces.yaml
â”‚       â””â”€â”€ staging/                    â† Staging overlay (replica=1, lower resources)
â”‚           â”œâ”€â”€ kustomization.yaml
â”‚           â””â”€â”€ namespaces.yaml
â”‚
â”œâ”€â”€ argo-rollouts/
â”‚   â”œâ”€â”€ canary-template.yaml            â† Canary rollout template (10â†’25â†’50â†’100%)
â”‚   â””â”€â”€ bluegreen-template.yaml         â† Blue-green rollout template
â”‚
â”œâ”€â”€ argo-events/
â”‚   â””â”€â”€ github-eventsource.yaml         â† GitHub webhook EventSource + Sensor â†’ Tekton
â”‚
â””â”€â”€ argo-workflows/
    â”œâ”€â”€ ci-build-workflow.yaml          â† CI pipeline (cloneâ†’testâ†’buildâ†’scanâ†’pushâ†’gitops update)
    â””â”€â”€ ml-training-workflow.yaml       â† ML model training workflow
```

---

## Installing GitOps Tools

Use `ci/jenkins/gitops.Jenkinsfile` to install any of the 12 GitOps tools onto the cluster.
Select `ACTION=INSTALL`, check the tools you want, and run. Each tool is installed from
`gitops/charts/<tool>/` using Helm. After install the tool URL and credentials are written
to `infra.env`.

| Tool | Namespace | Port | Credentials |
|---|---|---|---|
| ArgoCD | `argocd` | 8080 | admin / read from `argocd-initial-admin-secret` |
| Argo Rollouts | `argo-rollouts` | 3100 | â€” |
| Argo Workflows | `argo-workflows` | 2746 | admin / admin |
| Argo Events | `argo-events` | 7777 | â€” |
| ArgoCD Image Updater | `argocd` | 8080 | â€” |
| Flux CD | `flux-system` | 9292 | â€” |
| Flagger | `flagger` | 10080 | â€” |
| Weave GitOps | `weave-gitops` | 9001 | admin / admin |
| Sealed Secrets | `sealed-secrets` | 8080 | â€” |
| External Secrets | `external-secrets` | 8080 | â€” |
| vCluster | `vcluster` | 8443 | â€” |
| Gimlet | `gimlet` | 9000 | admin / gimlet |

---

## GitOps Deployment Pipeline

```mermaid
flowchart LR
    DEV([Developer]) -->|git push| REPO[Git Repository\nmain / release branch]

    REPO -->|Webhook| AE[Argo Events\nGitHub EventSource]
    AE -->|Trigger| AW[Argo Workflows\nCI Build Workflow]

    AW -->|clone â†’ test â†’ build â†’ scan| SCAN{Trivy scan\npassed?}
    SCAN -->|No| FAIL([Build failed])
    SCAN -->|Yes| REG[Harbor Registry\nimage:SHA]
    AW -->|Update image tag in gitops/| REPO

    REPO -->|Poll every 3 min| ACD[ArgoCD\nApp-of-Apps]
    ACD -->|Detect drift| SYNC{In Sync?}

    SYNC -->|Yes| IDLE([No action])
    SYNC -->|No| ROLLOUT[Argo Rollouts\nCanary / Blue-Green]

    ROLLOUT -->|Progressive rollout| K8S[Kubernetes Cluster]
    K8S -->|Metrics| PROM[Prometheus]
    PROM -->|Analysis pass / fail| ROLLOUT

    ROLLOUT -->|Pass| PROMOTE([Promote to 100%])
    ROLLOUT -->|Fail| ABORT([Automatic Rollback])
```

---

## ArgoCD â€” App-of-Apps Pattern

ArgoCD continuously reconciles the desired state in Git with the live state in Kubernetes.
ShopOS uses the App-of-Apps pattern: a single root `Application` (`app-of-apps.yaml`)
points at `gitops/argocd/applicationsets/` which contains an `ApplicationSet` that
generates one ArgoCD `Application` per service (154 total).

- Sync policy: `automated` with `selfHeal: true` and `prune: true`
- Projects: one `AppProject` per domain â€” scopes each team to their own namespace
- ApplicationSet: list generator covering all 154 services across 13 domains

```bash
# Bootstrap the app-of-apps (ArgoCD must already be installed)
kubectl apply -f gitops/argocd/app-of-apps.yaml -n argocd

# Port-forward ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:80

# CLI â€” sync a specific application
argocd app sync order-service

# Force hard refresh
argocd app get order-service --hard-refresh

# Manual rollback to previous version
argocd app rollback order-service
```

---

## Flux CD â€” Base / Overlay Pattern

Flux CD manages HelmReleases for all key services using a base/overlay pattern.
`flux/base/` holds the shared HelmRelease definitions; `clusters/production/` and
`clusters/staging/` patch replica counts and resource limits via Kustomize.

```bash
# Bootstrap Flux on a cluster
flux bootstrap github \
  --owner=your-org \
  --repository=shopos \
  --path=gitops/flux/clusters/production \
  --personal

# Check reconciliation status
flux get all -A

# Force a manual reconciliation
flux reconcile source git shopos
flux reconcile kustomization flux-system
```

---

## Argo Rollouts â€” Progressive Delivery

Argo Rollouts replaces standard `Deployment` objects with `Rollout` CRDs supporting
canary and blue-green strategies with automated Prometheus analysis at each step.

- Canary (default): 10% â†’ 25% â†’ 50% â†’ 75% â†’ 100% with pause between each step
- Blue-green: instant cutover with pre-promotion analysis gate
- Auto-rollback: triggered when error rate exceeds 1% or p99 latency breaches SLO

```bash
# Watch a rollout
kubectl argo rollouts get rollout order-service -n shopos-commerce --watch

# Manually promote a canary
kubectl argo rollouts promote order-service -n shopos-commerce

# Abort and roll back
kubectl argo rollouts abort order-service -n shopos-commerce
```

---

## Argo Events

Argo Events drives event-based automation. The GitHub EventSource listens for push
and pull_request events on the ShopOS repo and fires a Sensor that creates a Tekton
`PipelineRun` to trigger the CI build.

Key event sources:
- `github` â€” push/PR on `shopos` repo â†’ triggers CI pipeline
- `kafka` â€” `analytics.*` topic messages â†’ triggers data pipeline workflows
- `calendar` â€” nightly trigger for scheduled reconciliation jobs

---

## Argo Workflows â€” CI Pipeline

`argo-workflows/ci-build-workflow.yaml` is a `WorkflowTemplate` that runs a full
CI pipeline as a DAG:

| Step | Tool | Description |
|---|---|---|
| clone | alpine/git | Shallow clone at the target revision |
| test | language runtime | Run unit tests for the service |
| build | Kaniko | Build Docker image (no Docker socket needed) |
| scan | Trivy | Scan image for HIGH/CRITICAL CVEs |
| push | Kaniko | Push image to Harbor registry |
| update-image | alpine/git | Update image tag in `helm/charts/<service>/values.yaml` |

---

## Deployment Environments

| Environment | GitOps Engine | Strategy | Auto-Sync |
|---|---|---|---|
| `staging` | Flux + ArgoCD | Canary (2 steps), replica=1 | Yes |
| `production` | ArgoCD + Argo Rollouts | Canary (4 steps) with Prometheus analysis | Manual promote after 25% |

---

## References

- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Flux CD Documentation](https://fluxcd.io/docs/)
- [Argo Rollouts Documentation](https://argoproj.github.io/argo-rollouts/)
- [Argo Events Documentation](https://argoproj.github.io/argo-events/)
- [Argo Workflows Documentation](https://argoproj.github.io/argo-workflows/)
- [ShopOS CI Pipelines](../ci/README.md)
- [ShopOS Helm Charts](../helm/README.md)
