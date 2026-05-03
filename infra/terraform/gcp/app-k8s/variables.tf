variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "name" {
  type    = string
  default = "shopos-app-k8s"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "kubernetes_version" {
  type        = string
  default     = null
  description = "GKE release-channel version. null = STABLE channel."
}

variable "subnet_cidr" {
  type    = string
  default = "10.1.0.0/20"
}

variable "pods_cidr" {
  type    = string
  default = "10.2.0.0/16"
}

variable "services_cidr" {
  type    = string
  default = "10.3.0.0/20"
}

variable "master_cidr" {
  type    = string
  default = "172.16.0.0/28"
}
