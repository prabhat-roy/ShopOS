resource "google_compute_firewall" "master_to_nodes" {
  name    = "${var.name}-master-to-nodes"
  network = var.network_name

  allow {
    protocol = "tcp"
    ports    = ["443", "10250"]
  }

  source_ranges = [var.master_cidr]
  target_tags   = ["${var.name}-node"]
}
