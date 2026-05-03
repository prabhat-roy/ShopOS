resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name                                            = "${var.name}-vpc"
    Environment                                     = var.environment
    "kubernetes.io/cluster/${var.k8s_cluster_name}" = "shared"
  }
}
