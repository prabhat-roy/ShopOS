# Security — ShopOS

Defence-in-depth — every layer (secrets, identity, network, runtime, policy, supply chain,
SAST/DAST, posture) has at least one open-source control. All tooling is configuration-as-code.

> Production-ready as of 2026-05-02. See [`../docs/runbooks/incident-response.md`](../docs/runbooks/incident-response.md)
> for triage; the per-tool scope is documented in the table below.

---

## Layout

```
security/
├── vault/                  HA Raft Vault — Raft + KMS unseal + K8s/OIDC/JWT/AppRole auth +
│                           KV-per-domain + dynamic Postgres + AWS IAM + PKI int+root +
│                           Transit + TOTP + SSH-CA. Bootstrap scripts in vault/bootstrap/.
├── keycloak/               Customer + employee SSO realms
├── dex/                    OIDC federation broker (Azure AD / Okta / Google → K8s)
├── authentik/              Internal-tooling SSO (Grafana, ArgoCD, Harbor)
├── openfga/                Relationship-based authz (Zanzibar model)
├── cedar/                  AWS Cedar policies for resource-scoped authz
├── opa/                    OPA admission rego (data residency, cross-namespace)
├── kyverno/                9 baseline policies (privileged, runAsNonRoot, registries,
│                           Cosign verify, requests/limits, labels, no :latest, RO root, drop ALL caps)
├── kubewarden/             Wasm policy engine (custom domain rules)
├── falco/                  11 enterprise rules (shell, escape, miner, payment exfil, …)
├── tetragon/               eBPF policy enforcement (kill on violation)
├── tracee/                 eBPF event collection (forensics)
├── spire/                  SPIFFE workload identity (X.509 SVIDs)
├── cert-manager/           ClusterIssuers — Let's Encrypt prod+staging + Vault PKI
├── coraza-waf/             OWASP WAF (ModSecurity rules)
├── trivy/                  Container + IaC scanning (CI)
├── trivy-operator/         Continuous in-cluster vulnerability + misconfig + secret scan
├── grype/                  Secondary CVE scanner
├── checkov/                IaC scanning (Terraform + Helm + Dockerfile)
├── kics/                   Extended IaC scanning (K8s + Ansible)
├── terrascan/              Policy-as-code IaC scanning
├── semgrep/                SAST custom rules per language
├── sonarqube/              Code quality + security hotspots
├── kube-bench/             CIS Benchmark
├── kube-hunter/            K8s pen testing
├── kubescape/              NSA / CISA / MITRE ATT&CK posture
├── cosign/                 Sigstore image signing (keyless via OIDC)
├── rekor/                  Sigstore transparency log
├── fulcio/                 Sigstore CA
├── sigstore/               Policy Controller — admission-time Cosign verification
├── syft/                   SBOM generation (CycloneDX)
├── nuclei/                 CVE template scanning (DAST)
├── zap/                    OWASP ZAP DAST
├── teleport/               Zero-trust SSH + K8s exec + DB proxy
├── pomerium/               Identity-aware proxy for internal tools
├── external-secrets/       External Secrets Operator (Vault → K8s Secrets)
├── sealed-secrets/         GitOps-safe secrets (encrypted at rest)
├── cilium/                 L7 NetworkPolicies (HTTP-aware payment + auth filters)
├── network-policies/       (Symlink target) — actual policies in kubernetes/network-policies/
├── wazuh/                  SIEM + HIDS
├── suricata/               Network IDS/IPS
├── zeek/                   Traffic analysis (TLS fingerprinting)
├── openvas/                External attack-surface scanning
├── defectdojo/             Vulnerability finding aggregation
├── dependency-track/       SBOM ingestion + CVE correlation
├── gitguardian/            ggshield pre-commit hook (200+ secret providers)
└── wiz-cli/                Consolidated scan (Trivy + Grype + Syft + Checkov + Gitleaks)
```

---

## Defence-in-depth layers

```
1. Pre-commit          gitleaks, ggshield, hadolint, golangci-lint
2. CI (build)          Trivy fs, Grype, Syft (SBOM), Checkov, KICS, Terrascan, tfsec,
                       Semgrep, SonarQube, CodeQL (GH), Cosign sign, Rekor record
3. Registry            Harbor + Clair + Anchore policy
4. Admission           Kyverno (9 baseline), OPA (residency, cross-ns), Sigstore Policy
                       Controller (Cosign verify), Kubewarden (custom)
5. Network             Cilium L7, default-deny NetworkPolicies (all 22 namespaces),
                       Istio STRICT mTLS PeerAuth + AuthZ deny-all + named allows
6. Runtime             Falco (11 rules) + Tetragon (enforce) + Tracee (forensics) +
                       Trivy Operator (continuous), Kubescape posture
7. Identity            Keycloak (customer/employee), Dex (federation), Authentik (internal),
                       OpenFGA + Cedar (resource authz), SPIFFE/SPIRE (workload),
                       Pomerium + Teleport (zero-trust)
8. Secrets             Vault HA + dynamic creds + PKI + Transit + ESO + Sealed Secrets + SOPS
9. Observability       Wazuh SIEM, Suricata IDS, Zeek traffic, OpenVAS, OpenSearch security idx
10. Compliance         DefectDojo (find aggregation), Dependency-Track (SBOM CVE),
                       OpenSSF Scorecard, kube-bench (CIS), Kubescape (NSA)
```

---

## Vault bootstrap

```bash
# After Helm install:
kubectl exec -n infra vault-0 -- vault operator init -key-shares=5 -key-threshold=3
# (records 5 unseal keys + root token; store securely or use KMS auto-unseal)

# Then run the three bootstrap steps with VAULT_ADDR + VAULT_TOKEN set:
bash security/vault/bootstrap/01-auth-methods.sh    # K8s + OIDC + JWT + AppRole
bash security/vault/bootstrap/02-secret-engines.sh  # KV-per-domain + database + AWS + PKI + Transit + TOTP + SSH-CA
bash security/vault/bootstrap/03-policies-roles.sh  # bind policies to K8s ServiceAccounts
```

See [`vault/bootstrap/`](vault/bootstrap/) for the scripts.

---

## Apply admission + runtime policies

```bash
kubectl apply -f security/kyverno/policies/baseline-policies.yaml
kubectl apply -f security/sigstore/policy-controller.yaml
kubectl apply -f security/opa/policies/admission/
kubectl apply -f security/falco/rules/enterprise-rules.yaml
kubectl apply -f security/cilium/network-policies.yaml
kubectl apply -f networking/istio/security/peer-authentication.yaml
kubectl apply -f networking/istio/security/authorization-policies.yaml
kubectl apply -f kubernetes/network-policies/
```

---

## Run a consolidated scan

```bash
bash security/wiz-cli/scan.sh .   # Trivy + Grype + Syft + Checkov + Gitleaks
```

---

## References

- [Incident response runbook](../docs/runbooks/incident-response.md)
