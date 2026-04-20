output "jenkins_url" {
  value       = "http://${azurerm_public_ip.jenkins.ip_address}:8080"
  description = "Jenkins UI URL"
}

output "jenkins_public_ip" {
  value       = azurerm_public_ip.jenkins.ip_address
  description = "Jenkins server public IP"
}

output "ssh_command" {
  value       = "ssh -i <private_key> ${var.admin_username}@${azurerm_public_ip.jenkins.ip_address}"
  description = "SSH command to connect to Jenkins server"
}
