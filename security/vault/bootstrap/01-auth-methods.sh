#!/usr/bin/env bash
# Vault bootstrap step 1 — enable auth methods. Run once after `vault operator init`.
set -euo pipefail
: "${VAULT_ADDR:?}"
: "${VAULT_TOKEN:?}"

# Kubernetes auth — every pod authenticates with its ServiceAccount JWT
vault auth enable -path=kubernetes kubernetes || true
vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc" \
  token_reviewer_jwt="@/var/run/secrets/kubernetes.io/serviceaccount/token" \
  kubernetes_ca_cert=@/var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
  issuer="https://kubernetes.default.svc.cluster.local" \
  disable_iss_validation="false"

# AppRole — for non-K8s workloads (Jenkins agents, scheduled jobs)
vault auth enable -path=approle approle || true

# OIDC — engineer login via Keycloak/Authentik
vault auth enable -path=oidc oidc || true
vault write auth/oidc/config \
  oidc_discovery_url="https://keycloak.shopos.example.com/realms/shopos" \
  oidc_client_id="vault" \
  oidc_client_secret="${VAULT_OIDC_CLIENT_SECRET}" \
  default_role="engineer"

# JWT auth for CI (GitHub Actions OIDC, GitLab CI OIDC)
vault auth enable -path=jwt-github jwt || true
vault write auth/jwt-github/config \
  oidc_discovery_url="https://token.actions.githubusercontent.com" \
  bound_issuer="https://token.actions.githubusercontent.com"

# Userpass — break-glass admin only
vault auth enable userpass || true

echo "auth methods enabled"
