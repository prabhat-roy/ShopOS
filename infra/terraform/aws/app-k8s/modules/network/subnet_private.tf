resource "aws_subnet" "private" {
  count             = length(var.private_subnet_cidrs)
  vpc_id            = aws_vpc.this.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name                                            = "${var.name}-private-${count.index + 1}"
    Environment                                     = var.environment
    "kubernetes.io/cluster/${var.k8s_cluster_name}" = "owned"
    "kubernetes.io/role/internal-elb"               = "1"
  }
}
