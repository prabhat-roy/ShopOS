# Workflow Orchestration — ShopOS

Durable workflow execution for long-running business processes. ShopOS uses Temporal as
the primary workflow engine for business sagas, and Argo Workflows for infrastructure
and ML training DAGs.

---

## Directory Structure

```
workflow/
└── temporal/
    ├── server-config.yaml          ← Temporal server Helm values
    ├── namespaces.yaml             ← Temporal namespaces (shopos-prod, shopos-staging)
    └── workflow-mapping.md        ← Which services own which workflows
```

---

## Temporal

### What it does

Temporal provides fault-tolerant, stateful workflow execution. If a workflow worker crashes
mid-execution, Temporal replays the event history and resumes exactly where it stopped —
without duplicating side effects.

### Namespaces

| Temporal Namespace | Environment | Retention |
|---|---|---|
| `shopos-prod` | Production | 30 days |
| `shopos-staging` | Staging / QA | 7 days |

### Workflows

| Workflow | Owner Service | Trigger | Description |
|---|---|---|---|
| `CheckoutSaga` | `checkout-service` | Order placed | Orchestrates payment → inventory reserve → fulfilment → notification |
| `SubscriptionBillingCycle` | `subscription-billing-service` | Cron (monthly) | Charge, retry, dunning, suspension |
| `RefundSaga` | `return-refund-service` | Return approved | Reverse fulfilment → payment refund → loyalty adjustment |
| `KYCAMLWorkflow` | `kyc-aml-service` | User registration | Identity verification, document checks, screening |
| `SupplierOnboarding` | `supplier-portal-service` | New supplier signup | Document collection, approval, system provisioning |
| `DataSubjectRequest` | `gdpr-service` | GDPR request | Collect, package, and deliver personal data |
| `PurchaseOrderApproval` | `approval-workflow-service` | PO submitted | Multi-level approval with timeout and escalation |
| `InventoryReplenishment` | `warehouse-service` | Inventory low event | Auto-reorder from preferred vendor |

### Running Temporal Locally

```bash
# Start Temporal dev server (all-in-one)
temporal server start-dev

# Open Temporal UI
open http://localhost:8233

# List workflows in a namespace
temporal workflow list --namespace shopos-prod

# Show workflow history
temporal workflow show --workflow-id checkout-abc123 --namespace shopos-prod
```

### Deploying to Kubernetes

```bash
helm repo add temporal https://charts.temporal.io
helm upgrade --install temporal temporal/temporal \
  --namespace temporal --create-namespace \
  -f workflow/temporal/server-config.yaml

# Apply namespaces
kubectl apply -f workflow/temporal/namespaces.yaml
```

### Connecting a Worker

Workers register with the Temporal server and poll for tasks. Each microservice that owns
a workflow implements a worker process:

```go
// Go worker (checkout-service example)
c, _ := client.Dial(client.Options{HostPort: "temporal.temporal.svc:7233"})
w := worker.New(c, "checkout-queue", worker.Options{})
w.RegisterWorkflow(CheckoutSagaWorkflow)
w.RegisterActivity(ReserveInventoryActivity)
w.RegisterActivity(ProcessPaymentActivity)
w.Run(worker.InterruptCh())
```

---

## Argo Workflows

Used for infrastructure automation and ML training pipelines. Defined as `Workflow` and
`CronWorkflow` CRDs in the `argo` namespace.

Key workflows (see [gitops/argo-workflows/](../gitops/argo-workflows/)):

| Workflow | Schedule | Description |
|---|---|---|
| `ci-build` | On push (via Argo Events) | Build, test, scan, push image |
| `ml-training` | Weekly Sunday 01:00 | Retrain recommendation and price optimisation models |
| `data-export` | Daily 03:00 | Export analytics snapshots to ClickHouse |
| `velero-restore-test` | Monthly | Restore test from latest Velero backup |

---

## References

- [Temporal Documentation](https://docs.temporal.io/)
- [Argo Workflows](https://argoproj.github.io/argo-workflows/)
- [GitOps — Argo Workflows](../gitops/argo-workflows/)
- [Saga pattern — ADR 005](../docs/adr/005-saga-orchestration.md)
