variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "cluster_name" {
  type    = string
  default = "shopos-gke"
}

variable "kubernetes_version" {
  type        = string
  default     = null
  description = "GKE release channel version. null = latest stable."
}

variable "vpc_cidr" {
  type    = string
  default = "10.1.0.0/16"
}

variable "subnet_cidr" {
  type    = string
  default = "10.1.0.0/20"
}

variable "pods_cidr" {
  type        = string
  default     = "10.2.0.0/16"
  description = "Secondary IP range for pods"
}

variable "services_cidr" {
  type        = string
  default     = "10.3.0.0/20"
  description = "Secondary IP range for services"
}

variable "master_cidr" {
  type        = string
  default     = "172.16.0.0/28"
  description = "CIDR for GKE control plane (must be /28, not overlapping with VPC)"
}
