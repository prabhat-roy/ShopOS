output "jenkins_url" {
  value       = "http://${module.vm.public_ip}:8080"
  description = "Jenkins UI URL (default credentials: admin / admin — change on first login)"
}

output "jenkins_public_ip" {
  value = module.vm.public_ip
}

output "ssh_command" {
  value       = "ssh ubuntu@${module.vm.public_ip}"
  description = "Direct SSH from any cmd / shell — uses your default ~/.ssh/id_ed25519 (or id_rsa); no -i flag needed"
}
