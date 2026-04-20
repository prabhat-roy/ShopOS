resource "google_compute_address" "jenkins" {
  name   = "${var.name}-static-ip"
  region = var.region
}
