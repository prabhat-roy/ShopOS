path "secret/data/commerce/*" {
  capabilities = ["read", "list"]
}

path "secret/data/shared/kafka" {
  capabilities = ["read"]
}

path "secret/data/shared/postgres" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "pki/issue/commerce-services" {
  capabilities = ["create", "update"]
}
