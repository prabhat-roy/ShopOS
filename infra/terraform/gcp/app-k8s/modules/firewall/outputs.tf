output "firewall_names" {
  value = [
    google_compute_firewall.master_to_nodes.name,
    google_compute_firewall.node_to_node.name,
    google_compute_firewall.egress.name,
  ]
}
