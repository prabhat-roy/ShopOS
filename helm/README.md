# Helm Charts — ShopOS

All Helm charts live under `helm/` — no external `helm repo add` required at deploy time.
Every chart is vendored in-repo.

---

## Directory Structure

```
helm/
├── services/                       ← 303 per-service charts (one per microservice/frontend)
│   ├── api-gateway/
│   ├── order-service/
│   ├── cart-service/
│   ├── payment-service/
│
└── infra/                          ← Infrastructure tool charts (vendored, zero internet needed)
    ├── databases/                  ← 14 charts
    │   ├── clickhouse/             ← OLAP analytics
    │   ├── cockroachdb/            ← Distributed PostgreSQL
    │   ├── dragonfly/              ← Redis-compatible (high throughput)
    │   ├── eventstore/             ← Event sourcing DB
    │   ├── manticore-search/       ← Full-text search
    │   ├── meilisearch/            ← Typo-tolerant search
    │   ├── memcached/              ← Key-value cache
    │   ├── neo4j/                  ← Graph database
    │   ├── patroni/                ← HA PostgreSQL
    │   ├── pgbouncer/              ← Connection pooler
    │   ├── scylladb/               ← High-throughput time-series
    │   ├── surrealdb/              ← Multi-model DB
    │   ├── temporal/               ← Workflow orchestration
    │   ├── timescaledb/            ← Time-series
    │   ├── typesense/              ← Instant search
    │   ├── valkey/                 ← Redis-compatible (Linux Foundation)
    │   └── weaviate/               ← Vector database
    │
    ├── db-tools/                   ← 4 charts
    │   ├── pgadmin/
    │   ├── mongo-express/
    │   ├── redis-commander/
    │   └── bytebase/
    │
    ├── data/                       ← 9 charts (analytics, API management, lineage, quality)
    │   ├── superset/               ← BI dashboards
    │   ├── apisix/                 ← API gateway
    │   ├── hasura/                 ← GraphQL engine
    │   ├── tyk/                    ← API management
    │   ├── apache-atlas/           ← Data catalog
    │   ├── marquez/                ← Data lineage (server)
    │   ├── marquez-web/            ← Data lineage (UI)
    │   ├── great-expectations/     ← Data quality
    │   └── openlineage/            ← Lineage emitter
    │
    ├── gitops/                     ← 13 charts
    │   ├── argocd/
    │   ├── fluxcd/
    │   ├── argo-rollouts/
    │   ├── argo-workflows/
    │   ├── argo-events/
    │   ├── keda/
    │   ├── velero/
    │   ├── flagger/
    │   ├── atlantis/
    │   ├── infracost/
    │   ├── driftctl/
    │   ├── crossplane/
    │   └── backstage/
    │
    ├── messaging/                  ← 1 chart
    │   └── conduktor-gateway/      ← Kafka policy enforcement
    │
    ├── observability/              ← 6 charts
    │   ├── grafana-mimir/
    │   ├── kiali/
    │   ├── signoz/
    │   ├── netdata/
    │   ├── perses/
    │   └── victoria-logs/
    │
    ├── platform/                   ← 14 charts
    │   ├── botkube/                ← Slack K8s alerts
    │   ├── k8sgpt/                 ← AI K8s diagnostics
    │   ├── opencost/               ← Cost attribution
    │   ├── unleash/                ← Feature flags
    │   ├── cachet/                 ← Status page
    │   ├── grafana-oncall/         ← On-call scheduling
    │   ├── nomad/                  ← Workload orchestrator
    │   ├── atlantis/               ← Terraform GitOps
    │   ├── dagster/                ← Data orchestration
    │   ├── prefect/                ← Workflow orchestration
    │   ├── apache-camel/           ← Enterprise integration
    │   ├── pomerium/               ← Zero-trust access proxy
    │   ├── devpod/                 ← Dev environments
    │   └── score/                  ← Cloud-agnostic workload spec
    │
    ├── registry/                   ← 31 charts
    │   ├── harbor/
    │   ├── nexus/
    │   ├── gitea/
    │   ├── sonarqube/
    │   ├── chartmuseum/
    │   ├── zot/
    │   └── ... (25 more)
    │
    ├── security/                   ← 7 charts
    │   ├── teleport/               ← Zero-trust SSH + K8s access
    │   ├── defectdojo/             ← Vulnerability management
    │   ├── dependency-track/       ← SBOM + CVE correlation
    │   ├── wazuh/                  ← SIEM + HIDS
    │   ├── external-secrets/       ← Sync secrets from Vault/AWS SSM
    │   ├── sealed-secrets/         ← Encrypt secrets for GitOps
    │   └── pomerium/               ← Identity-aware proxy
    │
    └── testing/                    ← 4 charts
        ├── pact-broker/
        ├── wiremock/
        ├── k6-operator/
        └── toxiproxy/
```

### Tools deployed via direct YAML (not yet packaged into helm/infra/)

Configs live under their domain folders and are applied with `kubectl apply -f`:

| Tool | Path |
|---|---|
| Karpenter (NodePools) | [`../kubernetes/scaling/karpenter/`](../kubernetes/scaling/karpenter/) |
| VPA + Goldilocks | [`../kubernetes/scaling/vpa/`](../kubernetes/scaling/vpa/) |
| Longhorn / Rook-Ceph | [`../storage/`](../storage/) |
| Cilium L7 NetworkPolicies | [`../security/cilium/`](../security/cilium/) |
| Sigstore Policy Controller | [`../security/sigstore/`](../security/sigstore/) |
| Trivy Operator | [`../security/trivy-operator/`](../security/trivy-operator/) |
| Kubescape | [`../security/kubescape/`](../security/kubescape/) |
| Cert-manager ClusterIssuers | [`../security/cert-manager/`](../security/cert-manager/) |
| External Secrets Operator | [`../security/external-secrets/`](../security/external-secrets/) |
| Teleport | [`../security/teleport/`](../security/teleport/) |
| Varnish + Caddy + Anubis + ngrok-operator + MetalLB + Kube-VIP | [`../networking/`](../networking/) |
| Spin / SpinKube | [`../networking/edge/spin/`](../networking/edge/spin/) |
| Redpanda + Zilla | [`../messaging/`](../messaging/) |
| Quickwit + Parca + Komodor + Healthchecks + Perses | [`../observability/`](../observability/) |
| Kubecost | [`../finops/kubecost/`](../finops/kubecost/) |
| Cachet + Grafana Incident + Grafana OnCall | [`../incident/`](../incident/) |
| Coder + DevSpace + n8n + Windmill | [`../dev/`](../dev/) |
| Airbyte + Cube + Metabase + LakeFS + Dgraph + YugabyteDB | [`../data/`](../data/), [`../databases/`](../databases/) |
| Istio PeerAuthentication / AuthZ / DestinationRules / VirtualServices | [`../networking/istio/`](../networking/istio/) |

---

## Service Chart Layout

Every service chart under `helm/services/<name>/` follows the same layout:

```
helm/services/<service-name>/
├── Chart.yaml
├── values.yaml             ← Defaults (1 replica, debug logging, minimal resources)
├── values-dev.yaml         ← Local kind/minikube overrides
├── values-staging.yaml     ← Staging — production-like, reduced replicas
├── values-prod.yaml        ← Production — full replicas, HPA, PDB, strict resources
└── templates/
    ├── deployment.yaml
    ├── service.yaml
    ├── hpa.yaml
    ├── serviceaccount.yaml
    ├── configmap.yaml
    ├── servicemonitor.yaml ← Prometheus ServiceMonitor
    └── _helpers.tpl
```

---

## Environment Value Overrides

| File | Replicas | Resources | HPA | Logging |
|---|---|---|---|---|
| `values.yaml` | 1 | minimal | disabled | debug |
| `values-dev.yaml` | 1 | reduced, NodePort | disabled | debug |
| `values-staging.yaml` | 2 | production-like | disabled | info |
| `values-prod.yaml` | 3+ | full | enabled | warn |

---

## Common Commands

### Install a Service

```bash
helm install order-service helm/services/order-service \
  --namespace order-service \
  --create-namespace \
  -f helm/services/order-service/values-staging.yaml

helm install order-service helm/services/order-service \
  --namespace order-service \
  --create-namespace \
  -f helm/services/order-service/values-staging.yaml \
  --set image.tag=v1.4.2
```

### Upgrade a Service

```bash
helm upgrade order-service helm/services/order-service \
  --namespace order-service \
  -f helm/services/order-service/values-prod.yaml \
  --set image.tag=v1.5.0 \
  --atomic \
  --timeout 5m
```

### Install an Infrastructure Tool

```bash
# ClickHouse
helm upgrade --install clickhouse helm/infra/databases/clickhouse \
  --namespace databases --create-namespace

# Harbor registry
helm upgrade --install harbor helm/infra/registry/harbor \
  --namespace registry --create-namespace

# ArgoCD
helm upgrade --install argocd helm/infra/gitops/argocd \
  --namespace argocd --create-namespace
```

### Rollback

```bash
helm history order-service -n order-service
helm rollback order-service 2 -n order-service
```

### Deploy All Services (via Make)

```bash
make deploy-local          # All 303 services to their own namespaces
make deploy-svc SVC=order-service
make deploy-local TAG=v1.5.0
```

### Inspect and Debug

```bash
helm template order-service helm/services/order-service \
  -f helm/services/order-service/values-prod.yaml

helm lint helm/services/order-service

helm diff upgrade order-service helm/services/order-service \
  --namespace order-service \
  -f helm/services/order-service/values-prod.yaml \
  --set image.tag=v1.5.0
```

---

## Namespace Convention

Every service is deployed to its own namespace matching the service name:

```
order-service      → namespace: order-service
cart-service       → namespace: cart-service
payment-service    → namespace: payment-service
```

Infrastructure tools go into dedicated namespaces:

```
helm/infra/databases/*    → namespace: databases
helm/infra/registry/*     → namespace: registry
helm/infra/gitops/*       → namespace: argocd / flux-system
helm/infra/security/*     → namespace: security
helm/infra/observability/*→ namespace: observability
```

---

## Chart Conventions

- Every chart generates a dedicated `ServiceAccount` — never uses `default`
- `HorizontalPodAutoscaler` always templated, disabled by default, enabled via `autoscaling.enabled: true`
- All environment variables reference a `Secret` or `ConfigMap` — no hardcoded values
- Image: `{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}`
- Readiness and liveness probes always point to `/healthz`
- Prometheus `ServiceMonitor` included (enabled via `metrics.enabled: true`)
- Non-root user enforced via `securityContext.runAsNonRoot: true`

---

## Adding a new service chart

After scaffolding a new Go service with `bash scripts/bash/scaffold-service.sh`, copy a
sibling chart as a template:

```bash
cp -r helm/services/api-gateway helm/services/<new-service>
sed -i 's/api-gateway/<new-service>/g' helm/services/<new-service>/Chart.yaml \
       helm/services/<new-service>/values.yaml
# then add an entry to gitops/argocd/applicationsets/all-services.yaml
# and       gitops/flux/base/helm-releases.yaml
```

Charts include HPA, securityContext (runAsNonRoot, readOnlyRootFilesystem,
drop ALL caps), ServiceAccount, helpers, prometheus.io annotations, and probes
pointing at `/healthz`.

---

## ArgoCD Integration

Service charts are referenced by ArgoCD ApplicationSet in `gitops/argocd/applicationsets/all-services.yaml`:

```yaml
spec:
  source:
    repoURL: https://gitea.shopos.internal/shopos/shopos.git
    path: helm/services/{{service}}
    helm:
      valueFiles:
        - values-prod.yaml
  destination:
    namespace: '{{service}}'
```

Infrastructure tools are managed separately via individual ArgoCD Applications in `gitops/argocd/applications/`.

---

## Packaging and Publishing Charts

```bash
helm package helm/services/order-service --destination .helm-packages/
helm push .helm-packages/order-service-1.5.0.tgz oci://harbor.shopos.internal/charts

curl --data-binary "@.helm-packages/order-service-1.5.0.tgz" \
  http://chartmuseum.shopos.internal/api/charts
```

---

## References

- [Helm Documentation](https://helm.sh/docs/)
- [Helm Diff Plugin](https://github.com/databus23/helm-diff)
- [Harbor Registry](https://goharbor.io/)
- [ShopOS GitOps / ArgoCD](../gitops/README.md)
- [ShopOS CI Pipelines](../ci/README.md)
