variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "subnet_id" {
  type = string
}

variable "security_group_ids" {
  type = list(string)
}

variable "instance_profile_name" {
  type = string
}

variable "vm_size" {
  type = string
}

variable "disk_size_gb" {
  type = number
}

variable "key_name" {
  type = string
}

variable "user_data" {
  type    = string
  default = ""
}
