resource "google_compute_firewall" "ssh" {
  name    = "${var.name}-allow-ssh"
  network = var.network_name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = [var.ssh_source_cidr]
  target_tags   = ["${var.name}-server"]
}
