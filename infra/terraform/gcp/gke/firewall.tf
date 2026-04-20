resource "google_compute_firewall" "master_to_nodes" {
  name    = "${var.cluster_name}-master-to-nodes"
  network = google_compute_network.this.name

  allow {
    protocol = "tcp"
    ports    = ["443", "10250"]
  }

  source_ranges = [var.master_cidr]
  target_tags   = ["${var.cluster_name}-node"]
}

resource "google_compute_firewall" "node_to_node" {
  name    = "${var.cluster_name}-node-to-node"
  network = google_compute_network.this.name

  allow {
    protocol = "all"
  }

  source_tags = ["${var.cluster_name}-node"]
  target_tags = ["${var.cluster_name}-node"]
}

resource "google_compute_firewall" "egress" {
  name      = "${var.cluster_name}-egress"
  network   = google_compute_network.this.name
  direction = "EGRESS"

  allow {
    protocol = "all"
  }

  destination_ranges = ["0.0.0.0/0"]
}
