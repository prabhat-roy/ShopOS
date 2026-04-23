# Helm Charts — ShopOS

ShopOS uses Helm 3 to package, template, and deploy all 263 services onto Kubernetes.
Each service has its own self-contained chart under `helm/charts/` with per-environment
value overrides.

---

## Directory Structure

```
helm/
└── charts/                         ← 263 individual service charts
    ├── api-gateway/
    │   ├── Chart.yaml
    │   ├── values.yaml             ← Defaults (1 replica, debug logging, minimal resources)
    │   ├── values-dev.yaml         ← Local kind/minikube overrides
    │   ├── values-staging.yaml     ← Staging cluster — production-like, reduced replicas
    │   ├── values-prod.yaml        ← Production — full replicas, HPA, PDB, strict resources
    │   └── templates/
    │       ├── deployment.yaml
    │       ├── service.yaml
    │       ├── hpa.yaml
    │       ├── serviceaccount.yaml
    │       ├── configmap.yaml
    │       ├── servicemonitor.yaml ← Prometheus ServiceMonitor
    │       └── _helpers.tpl
    ├── order-service/
    ├── cart-service/
    ├── payment-service/
    ├── auth-service/
    └── ... (150 more charts)
```

---

## Environment Value Overrides

Each chart ships four value files:

| File | Replicas | Resources | HPA | Logging |
|---|---|---|---|---|
| `values.yaml` | 1 | minimal | disabled | debug |
| `values-dev.yaml` | 1 | reduced, NodePort | disabled | debug |
| `values-staging.yaml` | 2 | production-like | disabled | info |
| `values-prod.yaml` | 3+ | full | enabled | warn |

**Example diff for `order-service`:**

```yaml
# values.yaml (defaults)
replicaCount: 1
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits:   { cpu: 500m, memory: 512Mi }
autoscaling:
  enabled: false

# values-prod.yaml
replicaCount: 3
resources:
  requests: { cpu: 500m, memory: 512Mi }
  limits:   { cpu: 2000m, memory: 2Gi }
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
  targetCPUUtilizationPercentage: 60
```

---

## Common Commands

### Install a Service

```bash
# Install order-service in its own namespace
helm install order-service helm/charts/order-service \
  --namespace order-service \
  --create-namespace \
  -f helm/charts/order-service/values-staging.yaml

# Install with a specific image tag
helm install order-service helm/charts/order-service \
  --namespace order-service \
  --create-namespace \
  -f helm/charts/order-service/values-staging.yaml \
  --set image.tag=v1.4.2
```

### Upgrade a Service

```bash
helm upgrade order-service helm/charts/order-service \
  --namespace order-service \
  -f helm/charts/order-service/values-prod.yaml \
  --set image.tag=v1.5.0 \
  --atomic \
  --timeout 5m
```

### Rollback

```bash
# View release history
helm history order-service -n order-service

# Roll back to previous revision
helm rollback order-service 2 -n order-service

# Roll back to a specific revision
helm rollback order-service 5 -n order-service --wait
```

### Deploy All Services (via Make)

```bash
# Deploy all 263 services — each to its own namespace
make deploy-local

# Deploy a single service
make deploy-svc SVC=order-service

# Deploy with a custom image tag
make deploy-local TAG=v1.5.0
```

### Inspect and Debug

```bash
# Render templates without applying
helm template order-service helm/charts/order-service \
  -f helm/charts/order-service/values-prod.yaml

# Lint a chart
helm lint helm/charts/order-service

# Diff before upgrade (requires helm-diff plugin)
helm diff upgrade order-service helm/charts/order-service \
  --namespace order-service \
  -f helm/charts/order-service/values-prod.yaml \
  --set image.tag=v1.5.0
```

---

## Namespace Convention

Every service is deployed to its **own namespace** matching the service name:

```
order-service      → namespace: order-service
cart-service       → namespace: cart-service
payment-service    → namespace: payment-service
auth-service       → namespace: auth-service
```

Helm creates the namespace automatically via `--create-namespace`. This isolates services
at the network policy level and allows independent RBAC per service.

---

## Chart Conventions

- Every chart generates a dedicated `ServiceAccount` — never uses `default`
- `HorizontalPodAutoscaler` is always templated, disabled by default, enabled via `autoscaling.enabled: true`
- All environment variables reference a `Secret` or `ConfigMap` — no hardcoded values
- Image is always `{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}`
- Readiness and liveness probes always point to `/healthz`
- Prometheus `ServiceMonitor` is included (enabled via `metrics.enabled: true`)
- Non-root user enforced via `securityContext.runAsNonRoot: true`

---

## Packaging and Publishing Charts

```bash
# Package a chart
helm package helm/charts/order-service --destination .helm-packages/

# Push to Harbor OCI registry
helm push .helm-packages/order-service-1.5.0.tgz oci://harbor.shopos.internal/charts

# Push to ChartMuseum
curl --data-binary "@.helm-packages/order-service-1.5.0.tgz" \
  http://chartmuseum.shopos.internal/api/charts
```

---

## ArgoCD Integration

Each service chart is referenced by an ArgoCD `Application` in [gitops/argocd/](../gitops/argocd/).
ArgoCD watches the chart + values files in git and auto-syncs on changes.

```yaml
# gitops/argocd/applications/order-service.yaml (excerpt)
spec:
  source:
    repoURL: https://gitea.shopos.internal/shopos/shopos.git
    path: helm/charts/order-service
    helm:
      valueFiles:
        - values-prod.yaml
  destination:
    namespace: order-service
```

---

## References

- [Helm Documentation](https://helm.sh/docs/)
- [Helm Diff Plugin](https://github.com/databus23/helm-diff)
- [Harbor Registry](https://goharbor.io/)
- [ShopOS GitOps / ArgoCD](../gitops/README.md)
- [ShopOS CI Pipelines](../ci/README.md)
