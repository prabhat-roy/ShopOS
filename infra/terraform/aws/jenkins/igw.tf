resource "aws_internet_gateway" "jenkins" {
  vpc_id = aws_vpc.jenkins.id

  tags = {
    Name        = "${var.name}-igw"
    Environment = var.environment
  }
}
