resource "google_compute_address" "this" {
  name   = "${var.name}-public-ip"
  region = var.region
}
