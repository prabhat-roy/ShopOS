variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "vpc_cidr" {
  type = string
}

variable "public_subnet_cidrs" {
  type = list(string)
}

variable "private_subnet_cidrs" {
  type = list(string)
}

variable "availability_zones" {
  type = list(string)
}

variable "k8s_cluster_name" {
  type        = string
  description = "Cluster name to use in kubernetes.io/cluster/<name> tags"
}
