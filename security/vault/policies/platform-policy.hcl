# Platform services can read their own secrets
path "secret/data/platform/*" {
  capabilities = ["read", "list"]
}

path "secret/data/shared/*" {
  capabilities = ["read"]
}

# Allow token renewal
path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}

# PKI certificates for mTLS
path "pki/issue/platform-services" {
  capabilities = ["create", "update"]
}
