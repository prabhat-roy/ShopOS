output "jenkins_url" {
  value       = "http://${module.vm.public_ip}:8080"
  description = "Jenkins UI URL (default credentials: admin / admin — change on first login)"
}

output "jenkins_public_ip" {
  value = module.vm.public_ip
}

output "ssh_command" {
  value       = "ssh -i ~/.ssh/${var.key_name}.pem ubuntu@${module.vm.public_ip}"
  description = "Adjust the .pem path to wherever your AWS keypair lives"
}
