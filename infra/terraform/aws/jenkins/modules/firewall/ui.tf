resource "aws_vpc_security_group_ingress_rule" "ui" {
  security_group_id = aws_security_group.this.id
  ip_protocol       = "tcp"
  from_port         = 8080
  to_port           = 8080
  cidr_ipv4         = var.ui_source_cidr
  description       = "Jenkins UI"
}

resource "aws_vpc_security_group_ingress_rule" "https" {
  security_group_id = aws_security_group.this.id
  ip_protocol       = "tcp"
  from_port         = 443
  to_port           = 443
  cidr_ipv4         = var.ui_source_cidr
  description       = "HTTPS"
}
