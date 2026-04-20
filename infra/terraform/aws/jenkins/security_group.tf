data "http" "icanhazip" {
  url = "https://ipv4.icanhazip.com"
}

resource "aws_security_group" "jenkins" {
  name        = "${var.name}-sg"
  description = "Jenkins server security group"
  vpc_id      = aws_vpc.jenkins.id

  tags = {
    Name        = "${var.name}-sg"
    Environment = var.environment
  }
}

resource "aws_vpc_security_group_ingress_rule" "allow_ssh" {
  security_group_id = aws_security_group.jenkins.id
  ip_protocol       = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_ipv4         = "${chomp(data.http.icanhazip.response_body)}/32"
  description       = "Allow SSH access from my IP"
}

resource "aws_vpc_security_group_ingress_rule" "allow_jenkins_ui" {
  security_group_id = aws_security_group.jenkins.id
  ip_protocol       = "tcp"
  from_port         = 8080
  to_port           = 8080
  cidr_ipv4         = var.allowed_cidr
  description       = "Jenkins UI"
}

resource "aws_vpc_security_group_ingress_rule" "allow_https" {
  security_group_id = aws_security_group.jenkins.id
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_ipv4         = var.allowed_cidr
  description       = "HTTPS"
}

resource "aws_vpc_security_group_egress_rule" "allow_all_outbound" {
  security_group_id = aws_security_group.jenkins.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
  description       = "Allow all outbound traffic"
}
