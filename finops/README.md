# FinOps — ShopOS

Cost visibility and optimization for the entire stack. Pairs with right-sizing recommendations
from VPA + Goldilocks under [`kubernetes/scaling/`](../kubernetes/scaling/).

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [kubecost/](kubecost/) | Kubecost 2.4 | Multi-cluster cost allocation, idle resource recommendations, network cost tracking, forecasting |

## Related

- OpenCost (per-namespace attribution, complementary): [`kubernetes/opencost/`](../kubernetes/opencost/)
- Cloud cost on Terraform PRs (Infracost): [`infra/`](../infra/) — runs in Atlantis pipeline
- Right-sizing: [`kubernetes/scaling/vpa/`](../kubernetes/scaling/vpa/) (VPA + Goldilocks UI)
- Karpenter (replaces over-provisioned static node groups with on-demand right-sized): [`kubernetes/scaling/karpenter/`](../kubernetes/scaling/karpenter/)

## Usage

```bash
# Open Kubecost UI
kubectl port-forward -n monitoring svc/kubecost-cost-analyzer 9090:9090
# https://localhost:9090

# CSV cost report by namespace, last 7 days
curl 'http://kubecost-cost-analyzer.monitoring.svc:9090/model/allocation?window=7d&aggregate=namespace' | jq .

# Idle resource recommendations
curl 'http://kubecost-cost-analyzer.monitoring.svc:9090/model/savings/requestSizing'
```
