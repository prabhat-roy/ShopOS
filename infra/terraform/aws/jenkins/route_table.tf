resource "aws_route_table" "jenkins_public" {
  vpc_id = aws_vpc.jenkins.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.jenkins.id
  }

  tags = {
    Name        = "${var.name}-public-rt"
    Environment = var.environment
  }
}

resource "aws_route_table_association" "jenkins_public" {
  subnet_id      = aws_subnet.jenkins_public.id
  route_table_id = aws_route_table.jenkins_public.id
}
