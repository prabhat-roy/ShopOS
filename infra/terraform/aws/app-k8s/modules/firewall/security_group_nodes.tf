resource "aws_security_group" "nodes" {
  name        = "${var.name}-nodes-sg"
  description = "EKS worker nodes security group"
  vpc_id      = var.vpc_id

  ingress {
    description = "Node-to-node all traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
  }

  ingress {
    description     = "Control plane to node kubelets and pods"
    from_port       = 1025
    to_port         = 65535
    protocol        = "tcp"
    security_groups = [aws_security_group.cluster.id]
  }

  ingress {
    description     = "Control plane to node HTTPS"
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.cluster.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name                                            = "${var.name}-nodes-sg"
    Environment                                     = var.environment
    "kubernetes.io/cluster/${var.k8s_cluster_name}" = "owned"
  }
}
