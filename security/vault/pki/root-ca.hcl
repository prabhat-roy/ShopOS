# Vault PKI configuration for ShopOS mTLS
# Apply with: vault policy write pki-admin pki/root-ca.hcl

path "pki/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

path "pki_int/*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}
