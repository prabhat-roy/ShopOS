#!/usr/bin/env bash
# Vault bootstrap step 2 — enable secret engines and dynamic-secret roles.
set -euo pipefail

# KV v2 per domain (allows fine-grained policies)
for ns in platform identity catalog commerce supply-chain financial customer-experience \
          communications content analytics-ai b2b integrations affiliate marketplace \
          gamification developer-platform compliance sustainability events-ticketing \
          auction rental web; do
  vault secrets enable -path="kv/${ns}" -version=2 kv || true
done

# Database secret engine — dynamic Postgres credentials per service
vault secrets enable database || true
vault write database/config/postgres-primary \
  plugin_name=postgresql-database-plugin \
  allowed_roles="*" \
  connection_url="postgresql://{{username}}:{{password}}@postgres-primary.databases.svc:5432/postgres?sslmode=require" \
  username="vault_admin" password="${POSTGRES_VAULT_PASSWORD}"

# One dynamic role per domain (1h TTL, 24h max)
for domain in identity commerce catalog supply_chain financial b2b marketplace platform cx compliance sustainability events_ticketing auction rental gamification; do
  vault write database/roles/${domain}_app \
    db_name=postgres-primary \
    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';
                         GRANT \"${domain}_app\" TO \"{{name}}\";" \
    default_ttl="1h" max_ttl="24h"
done

# AWS dynamic IAM credentials (used by services pushing to S3/MinIO)
vault secrets enable aws || true
vault write aws/config/root \
  access_key="${AWS_VAULT_ACCESS_KEY}" \
  secret_key="${AWS_VAULT_SECRET_KEY}" \
  region=us-east-1

# PKI engine — short-lived service certificates (24h leaves; root in pki/)
vault secrets enable -path=pki pki || true
vault secrets tune -max-lease-ttl=87600h pki
vault write -field=certificate pki/root/generate/internal \
  common_name="ShopOS Root CA" ttl=87600h > /tmp/root-ca.pem

vault secrets enable -path=pki_int pki || true
vault secrets tune -max-lease-ttl=43800h pki_int
vault write pki_int/roles/service-cert \
  allowed_domains="svc.cluster.local,shopos.example.com" \
  allow_subdomains=true \
  max_ttl="24h"

# Transit engine — encryption-as-a-service for PII fields
vault secrets enable transit || true
vault write -f transit/keys/pii-encryption type=aes256-gcm96
vault write -f transit/keys/payment-tokens type=aes256-gcm96

# TOTP engine — for MFA bootstrap
vault secrets enable totp || true

# SSH engine — Boundary integration for human SSH
vault secrets enable -path=ssh-client-signer ssh || true
vault write ssh-client-signer/config/ca generate_signing_key=true

echo "secret engines enabled"
