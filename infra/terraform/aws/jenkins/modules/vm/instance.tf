# Bootstrapped via cloud-init (user_data). `terraform apply` returns once the VM is created.
# Bootstrap progress streams to /var/log/cloud-init-output.log on the VM.

resource "aws_instance" "this" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.vm_size
  subnet_id              = var.subnet_id
  key_name               = var.key_name
  vpc_security_group_ids = var.security_group_ids
  iam_instance_profile   = var.instance_profile_name
  user_data              = var.user_data

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
  }

  root_block_device {
    volume_type           = "gp3"
    volume_size           = var.disk_size_gb
    delete_on_termination = true
  }

  tags = {
    Name        = "${var.name}-server"
    Environment = var.environment
  }
}
