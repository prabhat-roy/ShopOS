resource "aws_key_pair" "this" {
  key_name   = "${var.name}-key"
  public_key = var.ssh_pub_key

  tags = {
    Name        = "${var.name}-key"
    Environment = var.environment
  }
}
