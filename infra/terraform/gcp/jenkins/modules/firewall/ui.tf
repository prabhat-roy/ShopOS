resource "google_compute_firewall" "ui" {
  name    = "${var.name}-allow-ui"
  network = var.network_name

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }

  source_ranges = [var.ui_source_cidr]
  target_tags   = ["${var.name}-server"]
}

resource "google_compute_firewall" "https" {
  name    = "${var.name}-allow-https"
  network = var.network_name

  allow {
    protocol = "tcp"
    ports    = ["443"]
  }

  source_ranges = [var.ui_source_cidr]
  target_tags   = ["${var.name}-server"]
}
