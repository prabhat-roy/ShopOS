datacenter = "dc1"
data_dir   = "/consul/data"
log_level  = "INFO"
node_name  = "shopos-consul"
server     = true
bootstrap  = true
ui_config {
  enabled = true
}
client_addr    = "0.0.0.0"
bind_addr      = "0.0.0.0"
advertise_addr = "127.0.0.1"

connect {
  enabled = true
}

performance {
  raft_multiplier = 1
}

telemetry {
  prometheus_retention_time = "60s"
  disable_hostname           = true
}
