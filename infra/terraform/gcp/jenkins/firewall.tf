data "http" "icanhazip" {
  url = "https://ipv4.icanhazip.com"
}

resource "google_compute_firewall" "jenkins_ssh" {
  name    = "${var.name}-allow-ssh"
  network = google_compute_network.jenkins.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["${chomp(data.http.icanhazip.response_body)}/32"]
  target_tags   = ["${var.name}-server"]
}

resource "google_compute_firewall" "jenkins_ui" {
  name    = "${var.name}-allow-ui"
  network = google_compute_network.jenkins.name

  allow {
    protocol = "tcp"
    ports    = ["8080", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["${var.name}-server"]
}

resource "google_compute_firewall" "jenkins_egress" {
  name      = "${var.name}-allow-egress"
  network   = google_compute_network.jenkins.name
  direction = "EGRESS"

  allow {
    protocol = "all"
  }

  destination_ranges = ["0.0.0.0/0"]
  target_tags        = ["${var.name}-server"]
}
