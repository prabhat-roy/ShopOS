resource "google_compute_subnetwork" "jenkins" {
  name          = "${var.name}-subnet"
  ip_cidr_range = var.subnet_cidr
  region        = var.region
  network       = google_compute_network.jenkins.id
}
