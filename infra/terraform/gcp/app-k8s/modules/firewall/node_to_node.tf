resource "google_compute_firewall" "node_to_node" {
  name    = "${var.name}-node-to-node"
  network = var.network_name

  allow {
    protocol = "all"
  }

  source_tags = ["${var.name}-node"]
  target_tags = ["${var.name}-node"]
}
