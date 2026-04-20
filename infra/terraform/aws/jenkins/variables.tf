variable "region" {
  type    = string
  default = "us-east-1"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "name" {
  type    = string
  default = "jenkins"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "public_subnet_cidr" {
  type    = string
  default = "10.0.1.0/24"
}

variable "instance_type" {
  type    = string
  default = "t3.xlarge"
}

variable "volume_size" {
  type    = number
  default = 200
}

variable "key_name" {
  type    = string
  default = "us-east-1"
}

variable "allowed_cidr" {
  type        = string
  default     = "0.0.0.0/0"
  description = "CIDR allowed to reach SSH and Jenkins UI — restrict to your IP in production"
}

variable "private_key_path" {
  type        = string
  description = "Absolute path to the .pem private key file used for SSH access to the Jenkins instance"
}
