resource "aws_security_group" "this" {
  name        = "${var.name}-sg"
  description = "Jenkins server security group"
  vpc_id      = var.vpc_id

  tags = {
    Name        = "${var.name}-sg"
    Environment = var.environment
  }
}
