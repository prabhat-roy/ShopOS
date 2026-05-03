# Nomad agent configuration for ShopOS

datacenter = "dc1"
data_dir   = "/opt/nomad/data"
log_level  = "INFO"

# Bind to all interfaces in Kubernetes
bind_addr = "0.0.0.0"

advertise {
  http = "{{ GetInterfaceIP \"eth0\" }}"
  rpc  = "{{ GetInterfaceIP \"eth0\" }}"
  serf = "{{ GetInterfaceIP \"eth0\" }}"
}

server {
  enabled          = true
  bootstrap_expect = 3

  server_join {
    retry_join     = ["nomad-0.nomad-headless", "nomad-1.nomad-headless", "nomad-2.nomad-headless"]
    retry_max      = 3
    retry_interval = "15s"
  }
}

client {
  enabled = true

  options = {
    "driver.raw_exec.enable" = "1"
    "docker.privileged.enabled" = "true"
  }
}

plugin "docker" {
  config {
    allow_privileged = false
    volumes {
      enabled = true
    }
  }
}

telemetry {
  collection_interval        = "1s"
  disable_hostname           = true
  prometheus_metrics         = true
  publish_allocation_metrics = true
  publish_node_metrics       = true
}

# ACL (enable in production)
acl {
  enabled = false
}
