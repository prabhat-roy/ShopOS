variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "identity_id" {
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

variable "user_data" {
  type    = string
  default = ""
}
