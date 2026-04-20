# Kubernetes Manifests — ShopOS

Raw Kubernetes manifests that sit beneath Helm: namespace declarations, RBAC, network policies,
resource quotas, pod disruption budgets, KEDA autoscalers, and Velero backup schedules.
These are applied once during cluster bootstrap and updated as the platform evolves.

---

## Directory Structure

```
kubernetes/
├── namespaces/                 ← 19 Namespace declarations
├── rbac/                       ← ServiceAccounts, Roles, ClusterRoles, Bindings
├── network-policies/           ← Default-deny + per-namespace allow rules
├── resource-quotas/            ← ResourceQuota + LimitRange per namespace
├── pod-disruption-budgets/     ← PDBs for all stateful and critical services
├── keda/                       ← KEDA ScaledObjects (Kafka + Redis triggers)
└── velero/                     ← Velero Schedule (daily backup to MinIO)
```

---

## Namespaces

| Namespace | Purpose |
|---|---|
| `platform` | API gateway, BFFs, platform services |
| `identity` | Auth, user, session, MFA services |
| `catalog` | Product catalog, search, pricing |
| `commerce` | Cart, checkout, order, payment |
| `supply-chain` | Vendor, warehouse, fulfillment |
| `financial` | Invoice, payout, accounting |
| `customer-experience` | Reviews, wishlist, support |
| `communications` | Notifications, email, SMS |
| `content` | Media, CMS, i18n |
| `analytics-ai` | ML, analytics, recommendations |
| `b2b` | Organisation, contracts, quotes |
| `integrations` | ERP, marketplace, CRM connectors |
| `affiliate` | Affiliate, referral, influencer |
| `messaging` | Kafka, RabbitMQ, NATS, ZooKeeper |
| `observability` | Prometheus, Grafana, Loki, Jaeger, OTel |
| `security` | Vault, Keycloak, Falco, Kyverno |
| `registry` | Harbor, MinIO, Nexus |
| `argocd` | ArgoCD, Argo Rollouts, Argo Workflows |
| `keda` | KEDA operator |

---

## RBAC

Each service has its own `ServiceAccount`. Roles follow least-privilege:

- **Platform services** (api-gateway, saga-orchestrator): read access to ConfigMaps and Secrets in their own namespace
- **ArgoCD**: cluster-scoped read + write to managed namespaces
- **Velero**: cluster-scoped backup/restore permissions
- **KEDA**: read ScaledObjects, write HPA resources
- **Falco / Tetragon**: node-level read access for runtime monitoring

Apply all RBAC:
```bash
kubectl apply -f kubernetes/rbac/
```

---

## Network Policies

The default posture is **deny-all ingress + deny-all egress** per namespace, with explicit
allow rules for known traffic flows:

| Policy | Scope | Allows |
|---|---|---|
| `default-deny` | All namespaces | Nothing (baseline) |
| `allow-intra-namespace` | Each namespace | Pods within the same namespace |
| `allow-prometheus-scrape` | All namespaces | `observability` namespace → `:metrics` |
| `allow-istio-sidecar` | All namespaces | Istiod control-plane traffic |
| `allow-ingress` | `platform` | Traefik → api-gateway port 8080 |
| `allow-kafka-egress` | All namespaces | Pods → `messaging` namespace port 9092 |
| `allow-postgres-egress` | Per namespace | Pods → their designated PG cluster |

Apply:
```bash
kubectl apply -f kubernetes/network-policies/
```

---

## Resource Quotas

Every namespace has a `ResourceQuota` and `LimitRange` to prevent runaway resource consumption.

Example quotas (production values):

| Namespace | CPU Request | Memory Request | CPU Limit | Memory Limit | Pods |
|---|---|---|---|---|---|
| `platform` | 20 | 40Gi | 40 | 80Gi | 100 |
| `commerce` | 30 | 60Gi | 60 | 120Gi | 150 |
| `analytics-ai` | 40 | 80Gi | 80 | 160Gi | 80 |
| `messaging` | 20 | 40Gi | 40 | 80Gi | 50 |
| `observability` | 15 | 30Gi | 30 | 60Gi | 60 |

Default `LimitRange` per container: request `100m` / `128Mi`, limit `2` / `4Gi`.

---

## Pod Disruption Budgets

Critical services have PDBs to guarantee availability during rolling updates and node drains.

| Service | Min Available |
|---|---|
| api-gateway | 2 |
| order-service | 2 |
| payment-service | 2 |
| kafka (each broker) | 2 |
| PostgreSQL (primary) | 1 |
| ArgoCD server | 1 |

---

## KEDA ScaledObjects

| ScaledObject | Trigger | Min | Max |
|---|---|---|---|
| `order-service` | Kafka `commerce.order.placed` lag | 2 | 20 |
| `notification-orchestrator` | Kafka `notification.*` lag | 1 | 15 |
| `fraud-detection-service` | Kafka `security.fraud.detected` lag | 2 | 10 |
| `cache-warming-service` | Redis list length | 1 | 5 |
| `email-service` | Kafka `notification.email.requested` lag | 1 | 10 |

---

## Velero Backup

A daily schedule backs up all namespaces to MinIO (`s3://velero-backups`):

```yaml
schedule: "0 2 * * *"   # 02:00 UTC daily
ttl: 720h               # retain for 30 days
storageLocation: minio
```

Manual backup:
```bash
velero backup create manual-$(date +%Y%m%d) --include-namespaces=commerce,platform,identity
```

Restore from latest:
```bash
velero restore create --from-backup $(velero backup get -o json | jq -r '.items[0].metadata.name')
```

---

## Applying Everything

```bash
# Apply in order (dependencies first)
kubectl apply -f kubernetes/namespaces/
kubectl apply -f kubernetes/rbac/
kubectl apply -f kubernetes/network-policies/
kubectl apply -f kubernetes/resource-quotas/
kubectl apply -f kubernetes/pod-disruption-budgets/
kubectl apply -f kubernetes/keda/
kubectl apply -f kubernetes/velero/
```

Or use the Makefile shortcut:
```bash
make k8s-bootstrap
```
