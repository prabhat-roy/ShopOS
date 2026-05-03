# Vault audit devices — file (mounted PV) + socket (forwarded to Loki via Fluent-bit).
# Apply with: vault audit enable file file_path=/vault/audit/audit.log
#             vault audit enable -path=loki socket address=fluent-bit.observability.svc:5170 socket_type=tcp

audit "file" {
  type = "file"
  options = {
    file_path = "/vault/audit/audit.log"
    log_raw   = "false"
    hmac_accessor = "true"
    mode      = "0600"
    format    = "json"
  }
}

audit "socket-loki" {
  type = "socket"
  options = {
    address     = "fluent-bit.observability.svc:5170"
    socket_type = "tcp"
    format      = "json"
  }
}
