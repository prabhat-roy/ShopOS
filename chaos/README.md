# Chaos Engineering — ShopOS

Chaos experiments that validate system resilience, test failure modes, and verify recovery
procedures. ShopOS uses two complementary frameworks: Chaos Mesh (Kubernetes-native) and
LitmusChaos (experiment marketplace).

---

## Directory Structure

```
chaos/
├── chaos-mesh/
│   ├── experiments/            ← 13 PodChaos / NetworkChaos / StressChaos CRDs
│   ├── workflows/              ← 2 ChaosWorkflow orchestrating multiple experiments
│   └── schedules/              ← 1 GameDay Schedule (monthly automated game day)
└── litmus/
    ├── experiments/            ← 5 ChaosEngine manifests
    └── workflows/              ← 2 Argo Workflow-based LitmusChaos runs
```

---

## Chaos Mesh Experiments

| Experiment | Kind | Target | Fault | Duration |
|---|---|---|---|---|
| `order-service-pod-kill` | PodChaos | `order-service` | Kill 1 pod | 1 min |
| `payment-service-latency` | NetworkChaos | `payment-service` | 500ms latency | 5 min |
| `kafka-network-partition` | NetworkChaos | Kafka brokers | Partition broker-0 | 2 min |
| `postgres-stress` | StressChaos | PostgreSQL | CPU 80% | 3 min |
| `identity-pod-kill` | PodChaos | `auth-service` | Kill all pods | 30 sec |
| `inventory-memory-hog` | StressChaos | `inventory-service` | Memory 2Gi | 2 min |
| `checkout-http-abort` | HTTPChaos | `checkout-service` | Abort 20% requests | 5 min |
| `redis-latency` | NetworkChaos | Redis | 200ms latency | 3 min |
| `dns-chaos-catalog` | DNSChaos | `catalog` namespace | Random DNS errors | 2 min |
| `node-drain` | NodeChaos | Random node | Cordon + drain | 5 min |
| `clock-skew` | TimeChaos | `financial` namespace | +30s clock skew | 3 min |
| `io-chaos-mongo` | IOChaos | MongoDB | 100ms IO latency | 3 min |
| `grpc-fault-injection` | NetworkChaos | `supply-chain` | Drop 10% gRPC | 2 min |

### Game Day Schedule

The monthly game day workflow runs a curated sequence of experiments against the staging
cluster every first Sunday at 02:00 UTC. It pauses 10 minutes between experiments to allow
monitoring dashboards to stabilise.

---

## LitmusChaos Experiments

| Experiment | Litmus Hub | Target | Description |
|---|---|---|---|
| `pod-delete-commerce` | generic/pod-delete | `commerce` namespace | Delete random pods every 30s |
| `container-kill-payment` | generic/container-kill | `payment-service` | Kill container (not pod) |
| `node-cpu-hog` | generic/node-cpu-hog | Worker nodes | 90% CPU for 4 minutes |
| `kafka-broker-pod-failure` | kafka/kafka-broker-pod-failure | `messaging` | Kill Kafka broker |
| `database-pod-delete` | generic/pod-delete | `postgres-*` | Delete primary pod |

---

## Running Experiments

### Chaos Mesh

```bash
# Apply a single experiment
kubectl apply -f chaos/chaos-mesh/experiments/order-service-pod-kill.yaml

# Check status
kubectl get podchaos -n commerce

# Stop experiment
kubectl delete podchaos order-service-pod-kill -n commerce

# Run the full game day workflow
kubectl apply -f chaos/chaos-mesh/schedules/game-day-schedule.yaml
```

### LitmusChaos

```bash
# Apply a ChaosEngine
kubectl apply -f chaos/litmus/experiments/pod-delete-commerce.yaml

# Check result
kubectl get chaosresult -n commerce

# LitmusChaos portal
kubectl port-forward svc/litmusportal-frontend-service 9091:9091 -n litmus
```

---

## Steady-State Hypotheses

| Metric | Threshold |
|---|---|
| order-service p95 latency | < 500ms |
| payment-service error rate | < 1% |
| Kafka consumer lag | < 10,000 messages |
| SLO burn rate (1h window) | < 5% |
| `/healthz` success rate | 100% |

---

## References

- [Chaos Mesh Documentation](https://chaos-mesh.org/docs/)
- [LitmusChaos Documentation](https://litmuschaos.io/docs)
- [SLO definitions](../observability/slo/)
- [Runbooks — incident response](../docs/runbooks/incident-response.md)
