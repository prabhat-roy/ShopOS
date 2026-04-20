# Networking — ShopOS

Helm charts and configurations for edge routing, service mesh, CNI, and service discovery.

## Directory Structure

```
networking/
├── traefik/            ← Traefik 3.1 — edge router, TLS termination, automatic service discovery
├── istio/              ← Istio service mesh — mTLS, traffic management, observability
├── cilium/             ← Cilium eBPF CNI — network policies, identity-aware filtering
├── consul/             ← Consul 1.19 — service discovery, health checking, K/V config
├── linkerd/            ← Linkerd — lightweight service mesh alternative
├── calico/             ← Calico CNI — network policies (alternative to Cilium)
├── kong/               ← Kong API Gateway (alternative to Traefik for API-level routing)
├── nginx-ingress/      ← NGINX Ingress Controller
├── haproxy-ingress/    ← HAProxy Ingress Controller
├── contour/            ← Contour Ingress (Envoy-based)
├── external-dns/       ← ExternalDNS — syncs K8s services to DNS providers
├── flannel/            ← Flannel CNI (simple overlay network)
├── antrea/             ← Antrea CNI (OVS-based, VMware)
└── weave-net/          ← Weave Net CNI
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| **Traefik** | 3.1 | Edge router — TLS termination, routing rules, automatic cert renewal |
| **Istio** | latest | Service mesh — mTLS between all pods, traffic policies, circuit breaking |
| **Cilium** | latest | eBPF CNI — fine-grained network policies, L7 filtering |
| **Consul** | 1.19 | Service discovery and health checking across cluster |

## Traffic Flow

```
Internet
  │ HTTPS (TLS)
  ▼
Traefik (edge)          ← cert-manager issues/renews TLS certs
  │ HTTP (inside cluster)
  ▼
API Gateway             ← JWT validation, rate limiting
  │ gRPC (mTLS via Istio)
  ▼
BFFs → Domain Services  ← Istio enforces mTLS on all pod-to-pod comms
```

## Traefik Configuration

- IngressRoute resources define routing rules per service
- Automatic TLS via ACME (Let's Encrypt) or internal CA
- Middlewares: rate limiting, headers, circuit breaking
- Dashboard available at `:8080/dashboard/`

## Istio Configuration

- Installed via Istio Operator in `istio-system` namespace
- All namespaces labelled `istio-injection: enabled`
- mTLS mode: `STRICT` — plaintext pod-to-pod traffic is rejected
- VirtualService and DestinationRule resources per service for canary traffic splitting

## Consul Configuration

- Service registry: all services register on startup with health check endpoint
- K/V store: feature flags and runtime config (alongside `config-service`)
- DNS interface: services resolve each other via `<service>.service.consul`

## Network Policies

Raw Kubernetes NetworkPolicy manifests are in `kubernetes/network-policies/`.
Cilium NetworkPolicy (CiliumNetworkPolicy CRDs) for L7-aware rules are in `networking/cilium/`.

Each service namespace only accepts ingress from its authorised callers — see [Domain Map](../docs/architecture/domain-map.md).

## References

- [Security Model](../docs/architecture/security-model.md)
- [System Overview](../docs/architecture/system-overview.md)
- [Kubernetes Network Policies](../kubernetes/network-policies/)
