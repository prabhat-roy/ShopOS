resource "aws_eip" "this" {
  instance = aws_instance.this.id
  domain   = "vpc"

  tags = {
    Name        = "${var.name}-eip"
    Environment = var.environment
  }
}
