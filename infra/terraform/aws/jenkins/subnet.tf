resource "aws_subnet" "jenkins_public" {
  vpc_id                  = aws_vpc.jenkins.id
  cidr_block              = var.public_subnet_cidr
  availability_zone       = "${var.region}a"
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.name}-public-subnet"
    Environment = var.environment
  }
}
