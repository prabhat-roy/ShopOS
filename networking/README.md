# Networking â€” ShopOS

Helm charts and configurations for edge routing, service mesh, CNI, and service discovery.

## Directory Structure

```
networking/
â”œâ”€â”€ traefik/            â† Traefik 3.1 â€” edge router, TLS termination, automatic service discovery
â”œâ”€â”€ istio/              â† Istio service mesh â€” mTLS, traffic management, observability
â”œâ”€â”€ cilium/             â† Cilium eBPF CNI â€” network policies, identity-aware filtering
â”œâ”€â”€ consul/             â† Consul 1.19 â€” service discovery, health checking, K/V config
â”œâ”€â”€ linkerd/            â† Linkerd â€” lightweight service mesh alternative
â”œâ”€â”€ calico/             â† Calico CNI â€” network policies (alternative to Cilium)
â”œâ”€â”€ kong/               â† Kong API Gateway (alternative to Traefik for API-level routing)
â”œâ”€â”€ nginx-ingress/      â† NGINX Ingress Controller
â”œâ”€â”€ haproxy-ingress/    â† HAProxy Ingress Controller
â”œâ”€â”€ contour/            â† Contour Ingress (Envoy-based)
â”œâ”€â”€ external-dns/       â† ExternalDNS â€” syncs K8s services to DNS providers
â”œâ”€â”€ flannel/            â† Flannel CNI (simple overlay network)
â”œâ”€â”€ antrea/             â† Antrea CNI (OVS-based, VMware)
â””â”€â”€ weave-net/          â† Weave Net CNI
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| Traefik | 3.1 | Edge router â€” TLS termination, routing rules, automatic cert renewal |
| Istio | latest | Service mesh â€” mTLS between all pods, traffic policies, circuit breaking |
| Cilium | latest | eBPF CNI â€” fine-grained network policies, L7 filtering |
| Consul | 1.19 | Service discovery and health checking across cluster |

## Traffic Flow

```
Internet
  â”‚ HTTPS (TLS)
  â–¼
Traefik (edge)          â† cert-manager issues/renews TLS certs
  â”‚ HTTP (inside cluster)
  â–¼
API Gateway             â† JWT validation, rate limiting
  â”‚ gRPC (mTLS via Istio)
  â–¼
BFFs â†’ Domain Services  â† Istio enforces mTLS on all pod-to-pod comms
```

## Traefik Configuration

- IngressRoute resources define routing rules per service
- Automatic TLS via ACME (Let's Encrypt) or internal CA
- Middlewares: rate limiting, headers, circuit breaking
- Dashboard available at `:8080/dashboard/`

## Istio Configuration

- Installed via Istio Operator in `istio-system` namespace
- All namespaces labelled `istio-injection: enabled`
- mTLS mode: `STRICT` â€” plaintext pod-to-pod traffic is rejected
- VirtualService and DestinationRule resources per service for canary traffic splitting

## Consul Configuration

- Service registry: all services register on startup with health check endpoint
- K/V store: feature flags and runtime config (alongside `config-service`)
- DNS interface: services resolve each other via `<service>.service.consul`

## Network Policies

Raw Kubernetes NetworkPolicy manifests are in `kubernetes/network-policies/`.
Cilium NetworkPolicy (CiliumNetworkPolicy CRDs) for L7-aware rules are in `networking/cilium/`.

Each service namespace only accepts ingress from its authorised callers â€” see [Domain Map](../docs/architecture/domain-map.md).

## References

- [Security Model](../docs/architecture/security-model.md)
- [System Overview](../docs/architecture/system-overview.md)
- [Kubernetes Network Policies](../kubernetes/network-policies/)
