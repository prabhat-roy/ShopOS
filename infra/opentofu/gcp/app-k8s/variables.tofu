variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "cluster_name" {
  type    = string
  default = "shopos-gke-dev"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "kubernetes_version" {
  type    = string
  default = null
}

variable "subnet_cidr" {
  type    = string
  default = "10.20.0.0/20"
}

variable "pods_cidr" {
  type    = string
  default = "10.21.0.0/16"
}

variable "services_cidr" {
  type    = string
  default = "10.22.0.0/20"
}
