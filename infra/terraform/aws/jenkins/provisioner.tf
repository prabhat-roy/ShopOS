resource "null_resource" "jenkins_setup" {
  depends_on = [aws_eip.jenkins]

  connection {
    type        = "ssh"
    host        = aws_eip.jenkins.public_ip
    user        = "ubuntu"
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
