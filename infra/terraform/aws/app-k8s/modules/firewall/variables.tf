variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "vpc_id" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "k8s_cluster_name" {
  type = string
}
