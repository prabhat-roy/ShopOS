# Bootstrapped via metadata_startup_script (cloud-init equivalent). OS Login disabled
# so direct `ssh ubuntu@<ip>` works using the caller's default ~/.ssh/id_ed25519.

resource "google_compute_instance" "this" {
  name         = "${var.name}-server"
  machine_type = var.vm_size
  zone         = var.zone
  tags         = ["${var.name}-server"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2604-lts-amd64"
      size  = var.disk_size_gb
      type  = "pd-ssd"
    }
  }

  network_interface {
    subnetwork = var.subnet_id
    access_config {
      nat_ip = google_compute_address.this.address
    }
  }

  service_account {
    email  = var.service_account_email
    scopes = ["cloud-platform"]
  }

  metadata = {
    enable-oslogin = "FALSE"
    ssh-keys       = "${var.admin_username}:${var.ssh_pub_key}"
  }

  metadata_startup_script = var.startup_script

  labels = {
    name        = "${var.name}-server"
    environment = var.environment
  }
}
