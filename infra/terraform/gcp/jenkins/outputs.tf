output "jenkins_url" {
  value       = "http://${google_compute_address.jenkins.address}:8080"
  description = "Jenkins UI URL"
}

output "jenkins_public_ip" {
  value       = google_compute_address.jenkins.address
  description = "Jenkins server public IP"
}

output "ssh_command" {
  value       = "ssh -i <private_key> ${var.ssh_user}@${google_compute_address.jenkins.address}"
  description = "SSH command to connect to Jenkins server"
}
