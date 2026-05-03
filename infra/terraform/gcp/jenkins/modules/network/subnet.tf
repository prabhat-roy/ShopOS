resource "google_compute_subnetwork" "this" {
  name                     = "${var.name}-subnet"
  ip_cidr_range            = var.subnet_cidr
  region                   = var.region
  network                  = google_compute_network.this.id
  private_ip_google_access = true
}
