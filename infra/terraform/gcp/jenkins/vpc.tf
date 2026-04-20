resource "google_compute_network" "jenkins" {
  name                    = "${var.name}-vpc"
  auto_create_subnetworks = false
}
