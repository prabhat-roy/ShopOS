# Security Model — ShopOS

ShopOS implements defence-in-depth across seven layers: edge, transport, identity, secrets, policy, runtime, and supply chain. Every layer is configured and automated — no manual security steps are required to deploy a service.

---

## Defence-in-Depth Overview

```
─Œ──────────────────────────────────────────────────────────────────────────────
│                       ShopOS Security Layers                                 │
│                                                                               │
│  ─Œ──────────────────────────────────────────────────────────────────────    │
│  │  Layer 7 — Supply Chain Security                                     │    │
│  │  Cosign image signing  ·  Syft SBOM  ·  Rekor transparency log      │    │
│  │  Fulcio certificate authority  ·  SLSA Level 2 provenance           │    │
│  ├──────────────────────────────────────────────────────────────────────¤    │
│  │  Layer 6 — Runtime Security                                          │    │
│  │  Falco syscall monitoring  ·  Tetragon eBPF enforcement             │    │
│  │  Tracee event collection  ·  Wazuh SIEM (log correlation + HIDS)    │    │
│  ├──────────────────────────────────────────────────────────────────────¤    │
│  │  Layer 5 — Policy Enforcement                                        │    │
│  │  OPA / Gatekeeper  ·  Kyverno  ·  Kubewarden (Wasm policies)        │    │
│  │  OpenFGA relationship-based authorisation                           │    │
│  ├──────────────────────────────────────────────────────────────────────¤    │
│  │  Layer 4 — Secrets Management                                        │    │
│  │  HashiCorp Vault (dynamic secrets)  ·  External Secrets Operator    │    │
│  │  Sealed Secrets (GitOps-safe)  ·  SOPS (file encryption)            │    │
│  ├──────────────────────────────────────────────────────────────────────¤    │
│  │  Layer 3 — Identity & Access                                         │    │
│  │  Keycloak (SSO / OIDC / OAuth 2.0)  ·  SPIFFE / SPIRE (workload)   │    │
│  │  Dex (OIDC federation)  ·  Authentik (IdP alternative)              │    │
│  ├──────────────────────────────────────────────────────────────────────¤    │
│  │  Layer 2 — Service-to-Service Transport                              │    │
│  │  Istio mTLS (all pod-to-pod)  ·  Cilium eBPF CNI                   │    │
│  │  Linkerd (alternative mesh)  ·  Calico (alternative CNI)            │    │
│  ├──────────────────────────────────────────────────────────────────────¤    │
│  │  Layer 1 — Edge / Perimeter                                          │    │
│  │  Traefik TLS termination  ·  Coraza WAF (OWASP Core Rule Set)       │    │
│  │  rate-limiter-service (Redis token bucket per IP and API key)       │    │
│  └──────────────────────────────────────────────────────────────────────˜    │
└──────────────────────────────────────────────────────────────────────────────˜
```

---

## Layer 1 — Edge Security

All external traffic enters the cluster through a single ingress point. No service is directly reachable from outside the cluster.

| Component | Role |
|---|---|
| Traefik 3.1 | TLS termination, automatic cert renewal via cert-manager (Let's Encrypt / internal CA) |
| Coraza WAF | OWASP Core Rule Set v4 — blocks SQLi, XSS, path traversal, RCE attempts at ingress |
| rate-limiter-service | Redis token bucket — per-IP and per-API-key rate limiting; configurable tiers |

Cert-manager provisions TLS certificates automatically — both from Let's Encrypt (public endpoints) and from the internal Vault PKI (cluster-internal services). Certificate rotation is automatic; no manual certificate management.

---

## Layer 2 — Transport Security (Service Mesh)

Every pod-to-pod connection is mutually authenticated. A compromised service cannot impersonate another service.

| Component | Role |
|---|---|
| Istio | mTLS between all pods; Citadel rotates certs every 24h; traffic policy enforcement |
| Cilium | eBPF CNI; L3/L4 network policies; identity-aware packet filtering; Hubble observability |
| Linkerd | Lightweight alternative mesh for services requiring minimal overhead |
| Calico | Alternative CNI for clusters where Cilium is not available |

Default-deny network policy: Every namespace starts with a `deny-all` ingress/egress policy. Explicit allow rules in `kubernetes/network-policies/` grant only the connections each service actually needs.

```yaml
# Example: order-service only accepts traffic from checkout-service and saga-orchestrator
ingress:
  - from:
      - namespaceSelector:
          matchLabels:
            kubernetes.io/metadata.name: checkout-service
      - namespaceSelector:
          matchLabels:
            kubernetes.io/metadata.name: saga-orchestrator
```

---

## Layer 3 — Identity & Access

### Human Identity — Keycloak 25.0
- SSO via OIDC / OAuth 2.0 for admin portal, developer tooling, Grafana, ArgoCD
- Realm per environment (dev, staging, prod) with distinct client configurations
- JWT tokens validated by `auth-service` before any gRPC call proceeds
- MFA enforced for admin roles (`mfa-service` — TOTP + WebAuthn)
- Token-exchange feature enabled — services can exchange tokens for downstream service tokens

### Workload Identity — SPIFFE / SPIRE
- Every service pod receives a SPIFFE SVID (X.509 certificate) via the SPIFFE CSI driver
- SVIDs are rotated every hour — short-lived, no long-lived service account tokens
- SPIRE server uses Kubernetes node attestation (verifies the pod's kubelet before issuing SVID)
- Istio uses SVID certs for mTLS handshakes — removes dependency on Kubernetes service accounts for mesh identity
- OIDC Discovery Provider enables SPIRE to federate with external systems

### Federation
- Dex federates OIDC across multiple identity providers (GitHub, LDAP, SAML)
- Authentik provides an alternative IdP with self-service user management

### External Access — API Keys
- Partner and integration access via `api-key-service` — keys hashed (bcrypt) in Postgres
- Keys scoped to specific gRPC methods and rate-limited by tier (bronze/silver/gold)
- Keys are rotatable without service restart via Vault dynamic secret lease renewal

---

## Layer 4 — Secrets Management

No secret is ever hardcoded, stored in plaintext in source code, or baked into a Docker image.

### HashiCorp Vault (HA Raft — 3 replicas)
- All database credentials, API keys, TLS private keys stored in Vault
- Services authenticate via Kubernetes Auth method — pod service account token → Vault token
- Dynamic secrets: Vault generates short-lived Postgres credentials per pod startup (15-minute TTL, auto-renewed)
- Vault PKI issues internal TLS certificates for cluster services
- Vault Transit engine provides envelope encryption for PII fields

### External Secrets Operator
- Reconciles Vault secrets into Kubernetes Secrets on a configurable refresh interval (default: 1m)
- Services read secrets from mounted K8s Secrets — no Vault SDK required in application code
- ESO SecretStores defined per namespace, scoped to minimum required paths

### Sealed Secrets
- GitOps-safe encrypted secrets checked into git
- Encrypted with cluster-specific Bitnami Sealed Secrets key — only the target cluster can decrypt
- Used for non-rotating bootstrap secrets that must live in the repo

### SOPS
- File-level encryption for secrets in configuration files (Ansible vars, Helm values)
- Integrates with Vault Transit or AWS KMS for key management

---

## Layer 5 — Policy Enforcement

Three complementary admission controllers enforce security posture at the Kubernetes API level. All policies run in `audit` mode on new clusters before switching to `enforce`.

### OPA / Gatekeeper (3 replicas, `gatekeeper-system`)
Admission controller validating every K8s resource before apply:
- All containers must run as non-root (`runAsNonRoot: true`)
- No `privileged: true` containers
- All images must come from the internal Harbor registry
- CPU and memory limits are required on every container
- `hostNetwork`, `hostPID`, `hostIPC` are blocked
- 60-second audit cycle scans existing resources for policy drift

### Kyverno (3 admission replicas, 2 background replicas)
Complementary policy engine with mutating capabilities:
- Mutating: Automatically injects `securityContext` defaults if absent
- Validating: Blocks images without valid Cosign signatures
- Generating: Creates default NetworkPolicy and ResourceQuota on namespace creation
- Webhook excludes `kube-system` and `kyverno` namespaces to avoid bootstrap deadlocks

### Kubewarden (CRDs + controller + defaults)
WebAssembly-based policy engine for fine-grained custom policies:
- Policies compiled to Wasm — language-agnostic (Rust, Go, Rego)
- `recommended-policies` installed in monitor mode (audit-only) initially
- Provides a fallback policy layer independent of OPA and Kyverno runtimes

### OpenFGA (Relationship-Based Authorisation)
- Used by `permission-service` to evaluate "can user X perform action Y on resource Z"
- ReBAC model: Users → Roles → Resources with inherited and contextual permissions
- Replaces flat RBAC for complex multi-tenant permission scenarios (B2B org hierarchies)

---

## Layer 6 — Runtime Security

| Component | Role |
|---|---|
| Falco | Syscall-level detection via eBPF driver; Falcosidekick forwards alerts to alertmanager, Slack, PagerDuty |
| Tetragon | eBPF enforcement — can kill processes violating network or file policy in real time |
| Tracee | eBPF event collection for forensic analysis and threat hunting |
| Wazuh | SIEM — log correlation, host intrusion detection (HIDS), file integrity monitoring, compliance dashboards |

### Key Falco rules
```yaml
- alert when any process writes to /etc/passwd or /etc/shadow
- alert when a container spawns a shell (/bin/sh, /bin/bash, /bin/zsh)
- alert when a process makes outbound connections to non-whitelisted external IPs
- alert when a process reads Kubernetes service account tokens
- alert when a binary is executed from /tmp or /dev/shm
```

### Wazuh SIEM integration
- Wazuh agents run as DaemonSets on every node
- Ingests Falco alerts, Kubernetes audit logs, container stdout logs
- Correlates events across nodes to detect lateral movement patterns
- Compliance dashboards for PCI-DSS, HIPAA, NIST
- Alerts forwarded to alertmanager → PagerDuty for P1/P2 severity

---

## Layer 7 — Supply Chain Security

The image promotion pipeline enforces that every image is scanned, signed, and attested before entering any environment.

```
Developer pushes code
  ←“
CI pipeline (Jenkins / Drone / Dagger)
  ├── Trivy scan           block on CRITICAL CVE
  ├── Grype scan           block on CRITICAL CVE (second opinion)
  ├── Semgrep SAST         block on HIGH security findings
  ├── OWASP Dep-Check      SCA for Java / Python / Node.js deps
  ├── Syft SBOM            generate CycloneDX SBOM per image
  ├── Cosign sign          sign image with Fulcio-issued cert (keyless)
  └── Rekor log            publish signature to transparency log
  ←“
Harbor registry
  ←“
ArgoCD deploys
  └── Kyverno verifies Cosign signature before admission
```

| Component | Role |
|---|---|
| Cosign (Sigstore) | Keyless image signing using Fulcio-issued short-lived certificate |
| Fulcio | Certificate authority — issues signing certs bound to OIDC identity (Keycloak) |
| Rekor | Append-only transparency log — all signatures publicly auditable |
| Syft | SBOM generation (CycloneDX + SPDX formats) per image at build time |
| CycloneDX | SBOM format uploaded to Dependency-Track for ongoing vulnerability correlation |
| SLSA Level 2 | Build provenance attestations generated by CI — signed and stored in Rekor |
| Trivy | Container + filesystem CVE scanner; blocks CRITICAL findings from being pushed |
| Grype | Second CVE scanner (Anchore data source) for defence-in-depth |
| OWASP Dep-Check | Software Composition Analysis for Java (Maven), Python (pip), Node.js (npm) |

---

## Scanning & DAST

| Tool | Type | Trigger |
|---|---|---|
| SonarQube | SAST — code quality + security rules (200+ security rules) | Every PR |
| Semgrep | SAST — custom security patterns, secrets detection | Every PR |
| Checkov | IaC scanning — Terraform, Helm, K8s manifests | Every PR touching infra |
| KICS | IaC scanning — broader rule set (500+ checks) | Every PR touching infra |
| OWASP ZAP | DAST — automated API fuzzing against live staging | Nightly |
| Nuclei | DAST — CVE template scanning against live endpoints | Nightly |
| kube-bench | CIS Kubernetes Benchmark (CIS 1.8) | Weekly cluster audit |
| kube-hunter | Kubernetes penetration testing (passive + active modes) | Weekly cluster audit |
| Kubescape | NSA/MITRE compliance scanning + network policy risk | Continuous in-cluster |

---

## GDPR & Compliance

| Service | Responsibility |
|---|---|
| `gdpr-service` | Handles data subject requests: access, erasure, portability (GDPR Art. 15/17/20) |
| `kyc-aml-service` | KYC checks at onboarding; AML transaction monitoring for financial compliance |
| `consent-management-service` | Tracks and enforces marketing consent per user per channel |
| `audit-service` | Append-only audit log for all privileged operations (Kafka → Postgres, 7-year retention) |

PII protection:
- All PII fields encrypted at rest using Vault Transit engine (envelope encryption)
- Encryption keys stored in Vault, not in the database or service configuration
- Field-level decryption only performed by the owning service — no plaintext PII in Kafka events

---

## Security Posture Summary

| Control | Status |
|---|---|
| Zero-trust network | âœ“ Istio mTLS + Cilium default-deny |
| Workload identity | âœ“ SPIFFE/SPIRE SVIDs, hourly rotation |
| No hardcoded secrets | âœ“ Vault dynamic secrets + ESO |
| Image signing enforced | âœ“ Cosign + Kyverno admission check |
| SBOM for every image | âœ“ Syft CycloneDX at build time |
| Runtime threat detection | âœ“ Falco + Tetragon + Wazuh SIEM |
| Policy-as-code | âœ“ OPA + Kyverno + Kubewarden |
| Compliance | âœ“ GDPR, PCI-DSS (Wazuh dashboards) |

---

## References

- [ADR-006: GitOps with ArgoCD](../adr/006-gitops-with-argocd.md)
- Security configs: `security/vault/`, `security/keycloak/`, `security/opa/`, `security/kyverno/`, `security/kubewarden/`, `security/falco/`, `security/spire/`
- Sigstore: `security/cosign/`, `security/rekor/`, `security/fulcio/`
- Network policies: `kubernetes/network-policies/`
- Kubescape config: `security/kubescape/`
- Wazuh config: `security/wazuh/` (via Helm install script)
