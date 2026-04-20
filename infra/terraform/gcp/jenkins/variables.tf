variable "project_id" {
  type        = string
  description = "GCP project ID"
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "zone" {
  type    = string
  default = "us-central1-a"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "name" {
  type    = string
  default = "jenkins"
}

variable "network_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "subnet_cidr" {
  type    = string
  default = "10.0.1.0/24"
}

variable "machine_type" {
  type    = string
  default = "n2-standard-4"
}

variable "disk_size_gb" {
  type    = number
  default = 200
}

variable "ssh_user" {
  type    = string
  default = "ubuntu"
}

variable "ssh_pub_key_path" {
  type        = string
  description = "Path to the SSH public key file for VM access"
}

variable "private_key_path" {
  type        = string
  description = "Path to the SSH private key file used by the provisioner"
}
