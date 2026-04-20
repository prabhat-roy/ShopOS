output "jenkins_url" {
  value       = "http://${aws_eip.jenkins.public_ip}:8080"
  description = "Jenkins UI URL"
}
