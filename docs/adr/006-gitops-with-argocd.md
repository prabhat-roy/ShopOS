# ADR-006: GitOps Deployment via ArgoCD App-of-Apps

Status: Accepted  
Date: 2024-02-01  
Deciders: Platform Architecture Team, DevOps Lead

---

## Context

With 154 services across 13 domains, we needed a deployment model that:
- Provides a single source of truth for what is running in each environment
- Enables automatic drift detection and reconciliation
- Supports progressive delivery (canary, blue-green) without custom scripting
- Scales to hundreds of services without becoming a manual process

We evaluated: Helm + kubectl (imperative), ArgoCD (declarative GitOps), Flux CD (pull-based GitOps).

---

## Decision

ArgoCD App-of-Apps pattern as the primary deployment mechanism, with Argo Rollouts for progressive delivery.

```
gitops/argocd/
â”œâ”€â”€ app-of-apps.yaml            â† Root application pointing to all AppProjects
â”œâ”€â”€ projects/                   â† 13 AppProjects (one per domain)
â”‚   â”œâ”€â”€ commerce-project.yaml
â”‚   â”œâ”€â”€ catalog-project.yaml
â”‚   â””â”€â”€ ...
â””â”€â”€ applicationsets/
    â””â”€â”€ all-services.yaml       â† ApplicationSet generating 154 Applications
```

Flux CD is maintained as an alternative GitOps engine with equivalent configuration in `gitops/flux/`.

Argo Rollouts defines canary and blue-green delivery strategies per service in `gitops/argo-rollouts/`.

---

## Rationale

1. Declarative â€” The desired state of every environment is fully expressed in git. Any manual `kubectl apply` creates drift that ArgoCD immediately detects and reports.
2. Automatic reconciliation â€” ArgoCD continuously compares live cluster state to git. Drift is auto-corrected or alerted, eliminating config drift over time.
3. ApplicationSet scalability â€” A single `ApplicationSet` with a list generator creates one ArgoCD `Application` per service. Adding a new service requires one line in the generator list, not a new pipeline.
4. Progressive delivery â€” Argo Rollouts canary strategy (10% â†’ 25% â†’ 50% â†’ 100%) with Prometheus analysis gates prevents bad deployments from reaching full traffic.
5. Audit trail â€” Every deployment is a git commit. Who deployed what, when, and what changed is in git history.

---

## Deployment Flow

```
git push (image tag update)
  â†’ ArgoCD detects diff
  â†’ Argo Rollouts starts canary
  â†’ 10% traffic â†’ Prometheus checks error rate
  â†’ If error_rate < 1%: promote to 25% â†’ 50% â†’ 100%
  â†’ If error_rate â‰¥ 1%: automatic rollback
```

---

## Consequences

Positive: Full audit trail in git; automatic drift correction; scalable to N services via ApplicationSet; canary deployments with automatic rollback; environment parity enforced structurally.

Negative: ArgoCD itself is a cluster dependency that must be highly available; the App-of-Apps structure adds one level of indirection to understand what is deployed.

Mitigations: ArgoCD is deployed with HA (3 replicas) and its own `Application` so it is self-managed; Weave GitOps provides a developer-friendly UI over ArgoCD for teams without deep K8s expertise; Flux CD serves as a hot-standby GitOps engine if ArgoCD is unavailable.
