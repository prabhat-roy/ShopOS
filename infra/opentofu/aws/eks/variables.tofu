variable "aws_region"   { default = "us-east-1" }
variable "cluster_name" { default = "shopos-eks" }
variable "environment"  { default = "prod" }
variable "vpc_id"       { type = string }
variable "subnet_ids"   { type = list(string) }
