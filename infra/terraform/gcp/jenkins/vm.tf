resource "google_compute_instance" "jenkins" {
  name         = "${var.name}-server"
  machine_type = var.machine_type
  zone         = var.zone
  tags         = ["${var.name}-server"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2404-lts-amd64"
      size  = var.disk_size_gb
      type  = "pd-ssd"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.jenkins.id
    access_config {
      nat_ip = google_compute_address.jenkins.address
    }
  }

  service_account {
    email  = google_service_account.jenkins.email
    scopes = ["cloud-platform"]
  }

  metadata = {
    startup-script = file("${path.root}/../../../../scripts/bash/jenkins-install.sh")
    ssh-keys       = "${var.ssh_user}:${file(var.ssh_pub_key_path)}"
  }

  labels = {
    name        = "${var.name}-server"
    environment = var.environment
  }
}

resource "null_resource" "jenkins_setup" {
  depends_on = [google_compute_instance.jenkins]

  connection {
    type        = "ssh"
    host        = google_compute_address.jenkins.address
    user        = var.ssh_user
    private_key = file(var.private_key_path)
    timeout     = "10m"
  }

  provisioner "remote-exec" {
    inline = [
      "echo 'Waiting for Jenkins user_data script to complete...'",
      "until [ -f /var/lib/jenkins/jenkins-setup-complete ]; do echo 'Still setting up...'; sleep 30; done",
      "echo 'Jenkins setup complete!'",
      "curl -s -o /dev/null -w 'Jenkins HTTP status: %%{http_code}' -u admin:admin http://localhost:8080/api/json",
    ]
  }
}
