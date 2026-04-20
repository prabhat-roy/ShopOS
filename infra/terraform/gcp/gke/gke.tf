resource "google_container_cluster" "this" {
  name     = var.cluster_name
  location = var.region

  # Autopilot = GCP equivalent of EKS Auto Mode (fully managed nodes)
  enable_autopilot = true

  min_master_version = var.kubernetes_version

  network    = google_compute_network.this.name
  subnetwork = google_compute_subnetwork.this.name

  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  private_cluster_config {
    enable_private_nodes    = true
    enable_private_endpoint = false
    master_ipv4_cidr_block  = var.master_cidr
  }

  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "0.0.0.0/0"
      display_name = "all"
    }
  }

  release_channel {
    channel = var.kubernetes_version == null ? "STABLE" : "UNSPECIFIED"
  }

  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  deletion_protection = false
}
