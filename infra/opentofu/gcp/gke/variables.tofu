variable "project_id"    { type = string }
variable "region"        { default = "us-central1" }
variable "cluster_name"  { default = "shopos-gke" }
variable "node_count"    { default = 3 }
variable "machine_type"  { default = "e2-standard-4" }
