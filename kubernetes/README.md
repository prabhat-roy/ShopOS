# Kubernetes — ShopOS

Raw Kubernetes manifests, namespace ops, autoscaling, backup, and per-service deployment
manifests. Helm and ArgoCD/Flux sit above these.

---

## Layout

```
kubernetes/
├── namespaces/                Namespace declarations (22 business + infra/monitoring/streaming/databases/...)
├── rbac/                      ServiceAccounts, Roles, ClusterRoles, RoleBindings
├── network-policies/          Default-deny + per-domain allow rules for all 22 namespaces
├── resource-quotas/           ResourceQuota + LimitRange per namespace
├── pod-disruption-budgets/    Explicit PDBs for tier-0/1 + Kyverno auto-PDB policy
├── keda/                      KEDA operator + Kafka/Redis ScaledObjects
├── scaling/
│   ├── karpenter/             Karpenter NodePools (general + spot-batch)
│   └── vpa/                   VPA + Goldilocks UI (recommendation-only)
├── velero/                    Velero install + backup schedules + restore-test CronJob
├── botkube/                   K8s event alerts to Slack
├── k8sgpt/                    AI-powered K8s diagnostics
├── opencost/                  Per-namespace cost attribution
└── manifests/                 Per-service raw K8s manifests (one subdir per domain)
                               — used as the source for Score generation and bare-K8s deploys
```

---

## Namespaces (22 + infra)

| Business namespaces | Infra namespaces |
|---|---|
| `platform`, `identity`, `catalog`, `commerce`, `supply-chain`, `financial`, `customer-experience`, `communications`, `content`, `analytics-ai`, `b2b`, `integrations`, `affiliate`, `marketplace`, `gamification`, `developer-platform`, `compliance`, `sustainability`, `events-ticketing`, `auction`, `rental`, `shopos-web` | `infra`, `monitoring`, `streaming`, `databases`, `temporal-system`, `flink-system`, `velero`, `keda`, `karpenter`, `longhorn-system`, `rook-ceph`, `trivy-system`, `kubescape`, `cosign-system`, `external-secrets`, `cert-manager`, `kyverno`, `argocd`, `istio-system`, `metallb-system`, `cilium`, `coder`, `teleport`, `spin-system`, `goldilocks` |

---

## Apply order (cluster bootstrap)

```bash
# 1. Namespaces + RBAC + quotas + policies (no workload yet)
kubectl apply -f kubernetes/namespaces/
kubectl apply -f kubernetes/rbac/
kubectl apply -f kubernetes/resource-quotas/
kubectl apply -f kubernetes/network-policies/

# 2. Admission policies (Kyverno + OPA + Sigstore Policy)
kubectl apply -f security/kyverno/policies/baseline-policies.yaml
kubectl apply -f security/sigstore/policy-controller.yaml

# 3. Mesh + identity (Istio + cert-manager + Vault + ESO)
kubectl apply -f networking/istio/
kubectl apply -f security/cert-manager/cluster-issuers.yaml
helm upgrade --install vault security/vault/charts/vault -n infra
bash security/vault/bootstrap/01-auth-methods.sh
bash security/vault/bootstrap/02-secret-engines.sh
bash security/vault/bootstrap/03-policies-roles.sh
kubectl apply -f security/external-secrets/external-secrets.yaml

# 4. Autoscaling + storage + backup
kubectl apply -f kubernetes/keda/
kubectl apply -f kubernetes/scaling/karpenter/karpenter.yaml
kubectl apply -f kubernetes/scaling/vpa/vpa.yaml
kubectl apply -f storage/longhorn/longhorn.yaml
kubectl apply -f kubernetes/velero/velero-install.yaml
kubectl apply -f kubernetes/velero/restore-tests.yaml

# 5. PDBs (after workloads exist)
kubectl apply -f kubernetes/pod-disruption-budgets/

# 6. ArgoCD bootstraps everything else from gitops/
helm upgrade --install argocd helm/infra/gitops/argocd/argo-cd -n argocd --create-namespace
kubectl apply -f gitops/argocd/app-of-apps.yaml
```

---

## Resource quotas + LimitRange

Every namespace has a `ResourceQuota` (CPU/memory request/limit + pod count) and a default
`LimitRange` (per-container request 100m/128Mi, limit 2/4Gi). Workloads without explicit
requests/limits are blocked by Kyverno (`require-requests-limits` policy).

---

## Pod Disruption Budgets

Tier-0 services have `minAvailable: 2` PDBs in [`pod-disruption-budgets/pdbs.yaml`](pod-disruption-budgets/pdbs.yaml).
Any Deployment with `replicas >= 2` that lacks an explicit PDB gets a permissive
`maxUnavailable: 25%` PDB auto-created via Kyverno
([`pod-disruption-budgets/default-pdb-policy.yaml`](pod-disruption-budgets/default-pdb-policy.yaml)).

---

## KEDA ScaledObjects

| ScaledObject | Trigger | Min | Max |
|---|---|---|---|
| `order-service` | Kafka lag on `commerce.order.placed` | 2 | 20 |
| `notification-orchestrator` | Kafka lag on `notification.*` | 1 | 15 |
| `fraud-detection-service` | Kafka lag on `security.fraud.detected` | 2 | 10 |
| `cache-warming-service` | Redis list length | 1 | 5 |
| `email-service` | Kafka lag on `notification.email.requested` | 1 | 10 |

---

## Karpenter + VPA + Goldilocks

- **Karpenter** ([`scaling/karpenter/`](scaling/karpenter/)) — provisions right-sized EC2 nodes on demand.
  NodePools: `general` (Graviton AMD64+ARM64 spot+on-demand) and `spot-batch` (taint=workload=batch).
- **VPA** ([`scaling/vpa/`](scaling/vpa/)) — recommendation-only mode (does NOT auto-evict pods).
- **Goldilocks** UI surfaces VPA recommendations for review.

---

## Velero backups

| Schedule | Cadence | Namespaces | Retention |
|---|---|---|---|
| `tier0-business` | every 6h | commerce, identity, catalog, financial, supply-chain | 14 days |
| `tier1-business` | daily 02:00 | 15 secondary domains + shopos-web | 30 days |
| `stateful-infra` | every 4h | databases, streaming, infra, temporal-system, flink-system | 30 days |
| `observability` | daily 03:30 | monitoring | 7 days |
| `weekly-dr-replication` | weekly Sun 04:00 | tier-0 + databases + streaming | 90 days, cross-region S3 |

A weekly `CronJob` ([`velero/restore-tests.yaml`](velero/restore-tests.yaml)) restores the
latest tier-1 backup to a temp namespace and asserts pods become ready — failure pages oncall.

---

## Per-service raw manifests

[`manifests/<domain>/<service>/`](manifests/) contains `deployment.yaml`, `service.yaml`,
`hpa.yaml`, `configmap.yaml`, `serviceaccount.yaml` for every service. These are the source
for [`dev/score/`](../dev/score/) workload specs and for bare-K8s installs that don't use Helm.

---

## Related

- Helm charts: [`../helm/`](../helm/)
- GitOps: [`../gitops/`](../gitops/)
- Backup runbook: [`../docs/runbooks/`](../docs/runbooks/)
