variable "cluster_name"    { default = "shopos-aks" }
variable "location"        { default = "East US" }
variable "resource_group"  { type = string }
variable "node_count"      { default = 3 }
variable "vm_size"         { default = "Standard_D4s_v3" }
