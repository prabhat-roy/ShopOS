# Networking — ShopOS

Edge routing, service mesh, CNI, service discovery, edge functions, and anti-bot protection.

## Layout

```
networking/
├── traefik/             Traefik 3.1 — primary edge router (TLS, routing, automatic cert renewal)
├── caddy/               Caddy — alternative ingress for non-prod / preview environments
├── istio/               Istio service mesh
│   ├── security/        STRICT mTLS PeerAuthentication + AuthorizationPolicies (deny-all + named allows)
│   └── traffic/         DestinationRules (outlier detection, conn pools), VirtualServices (canary, retries)
├── cilium/              Cilium eBPF CNI — L3/L4 + Hubble; L7 NetworkPolicies in security/cilium/
├── consul/              Service discovery, health checking, K/V config
├── linkerd/             Lightweight service mesh alternative
├── calico/              CNI alternative (NetworkPolicy)
├── kong/                API Gateway alternative to Traefik
├── nginx-ingress/       NGINX Ingress Controller
├── haproxy-ingress/     HAProxy Ingress Controller
├── contour/             Contour Ingress (Envoy-based)
├── external-dns/        Syncs K8s services to DNS providers (Route53/Cloudflare)
├── flannel/             Flannel CNI (simple overlay)
├── antrea/              Antrea CNI (OVS-based)
├── weave-net/           Weave Net CNI
├── cdn/                 Varnish HTTP cache + cdn-purge.sh helper (Cloudflare/Fastly)
├── anubis/              Anubis — proof-of-work anti-bot / anti-AI scraper for storefront
├── ngrok-operator/      ngrok Kubernetes Operator — public ingress for PR-preview environments
├── metallb/             MetalLB — bare-metal LoadBalancer for on-prem
├── kube-vip/            Kube-VIP — HA control-plane VIP for bare-metal
└── edge/
    └── spin/            Fermyon Spin / SpinKube — Wasm serverless edge functions
                         (storefront geo-IP greeting, A/B variant selection)
```

## Traffic flow

```
Internet (HTTPS, TLS)
  │
  ▼
Cloudflare/Fastly CDN  (purged via networking/cdn/cdn-purge.sh)
  │
  ▼
Anubis (PoW anti-bot)  ──► Varnish (HTTP cache)
  │
  ▼
Istio Gateway (TLS termination, cert-manager-issued cert)
  │
  ▼
api-gateway / web-bff / mobile-bff / partner-bff (JWT validation, rate limit)
  │ gRPC over mTLS (Istio STRICT PeerAuthentication)
  ▼
Domain services
  │
  ▼
Postgres / Redis / Kafka  (Cilium L7 NetworkPolicies on payment + auth ingress)
```

## Istio configuration (production-ready)

- **PeerAuthentication**: mesh-wide `STRICT` mTLS in [`istio/security/peer-authentication.yaml`](istio/security/peer-authentication.yaml) — every pod-to-pod call is mTLS
- **AuthorizationPolicies**: deny-all in [`istio/security/authorization-policies.yaml`](istio/security/authorization-policies.yaml) + named allows (BFF→cart, checkout→payment, order→fulfillment)
- **DestinationRules**: outlier detection + connection pools for tier-0 services in [`istio/traffic/destination-rules.yaml`](istio/traffic/destination-rules.yaml)
- **VirtualServices**: canary traffic-shifting + retries + edge gateway routing in [`istio/traffic/virtual-services.yaml`](istio/traffic/virtual-services.yaml)
- **Kiali**: topology UI in [`../observability/kiali/`](../observability/kiali/)

## Network policies

| Layer | Tool | Path |
|---|---|---|
| L3/L4 default-deny per namespace | Kubernetes NetworkPolicy | [`../kubernetes/network-policies/`](../kubernetes/network-policies/) |
| L7 HTTP-aware filtering | Cilium CiliumNetworkPolicy | [`../security/cilium/`](../security/cilium/) |
| Mesh-level authz | Istio AuthorizationPolicy | [`istio/security/`](istio/security/) |

## Edge compute

[`edge/spin/`](edge/spin/) deploys Fermyon Spin via the SpinKube operator, running compiled Wasm
modules as `SpinApp` CRDs. Used for low-latency storefront personalization that doesn't justify a
full microservice (geo-IP greeting, A/B variant selection, header rewriting).

## References

- [Security model](../docs/architecture/security-model.md)
- [System overview](../docs/architecture/system-overview.md)
- [Domain map](../docs/architecture/domain-map.md)
