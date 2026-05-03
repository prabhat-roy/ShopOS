# HA Vault server config — Raft storage, K8s service-account auth, audit to file + Loki sink.
ui = true
disable_mlock = true
api_addr     = "https://vault.shopos.svc.cluster.local:8200"
cluster_addr = "https://vault.shopos.svc.cluster.local:8201"

storage "raft" {
  path    = "/vault/data"
  node_id = "${HOSTNAME}"

  retry_join {
    leader_api_addr = "https://vault-0.vault-internal:8200"
  }
  retry_join {
    leader_api_addr = "https://vault-1.vault-internal:8200"
  }
  retry_join {
    leader_api_addr = "https://vault-2.vault-internal:8200"
  }
}

listener "tcp" {
  address       = "0.0.0.0:8200"
  cluster_address = "0.0.0.0:8201"
  tls_cert_file = "/vault/tls/tls.crt"
  tls_key_file  = "/vault/tls/tls.key"
  tls_min_version = "tls13"
}

seal "awskms" {
  region     = "us-east-1"
  kms_key_id = "alias/shopos-vault-unseal"
}

telemetry {
  prometheus_retention_time = "24h"
  disable_hostname = true
}

log_level = "info"
