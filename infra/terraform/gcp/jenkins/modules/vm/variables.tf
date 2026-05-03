variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "region" {
  type = string
}

variable "zone" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "service_account_email" {
  type = string
}

variable "vm_size" {
  type = string
}

variable "disk_size_gb" {
  type = number
}

variable "admin_username" {
  type = string
}

variable "ssh_pub_key" {
  type        = string
  description = "Public SSH key content (NOT a path)"
}

variable "startup_script" {
  type    = string
  default = ""
}
