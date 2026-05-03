resource "google_compute_firewall" "egress" {
  name      = "${var.name}-allow-egress"
  network   = var.network_name
  direction = "EGRESS"

  allow {
    protocol = "all"
  }

  destination_ranges = ["0.0.0.0/0"]
  target_tags        = ["${var.name}-server"]
}
