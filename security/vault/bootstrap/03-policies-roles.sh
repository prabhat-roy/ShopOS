#!/usr/bin/env bash
# Vault bootstrap step 3 — bind policies to Kubernetes ServiceAccounts.
set -euo pipefail

# Per-domain policy: read kv/<domain>/* and request DB creds for own role
for domain in identity commerce catalog supply-chain financial b2b marketplace platform \
              customer-experience compliance sustainability events-ticketing auction \
              rental gamification communications content integrations affiliate \
              developer-platform analytics-ai; do
  policy_name="${domain//-/_}_policy"
  cat <<EOF | vault policy write "${policy_name}" -
path "kv/${domain}/data/*" { capabilities = ["read","list"] }
path "kv/${domain}/metadata/*" { capabilities = ["read","list"] }
path "database/creds/${domain//-/_}_app" { capabilities = ["read"] }
path "transit/encrypt/pii-encryption" { capabilities = ["update"] }
path "transit/decrypt/pii-encryption" { capabilities = ["update"] }
path "pki_int/issue/service-cert" { capabilities = ["update"] }
EOF

  # Bind to all ServiceAccounts in that namespace
  vault write auth/kubernetes/role/${domain} \
    bound_service_account_names="*" \
    bound_service_account_namespaces="${domain}" \
    policies="${policy_name}" \
    ttl="1h"
done

# Engineer OIDC role — full read on most paths, no destroy
cat <<'EOF' | vault policy write engineer -
path "kv/+/data/*" { capabilities = ["read","list"] }
path "kv/+/metadata/*" { capabilities = ["read","list"] }
path "sys/health" { capabilities = ["read"] }
EOF
vault write auth/oidc/role/engineer \
  bound_audiences="vault" \
  user_claim="sub" \
  policies="engineer" \
  ttl="8h"

# Admin policy — break-glass only
cat <<'EOF' | vault policy write admin -
path "*" { capabilities = ["create","read","update","delete","list","sudo"] }
EOF

echo "policies and roles bound"
