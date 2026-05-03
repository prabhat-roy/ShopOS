module "aks" {
  source = "./modules/aks"

  name                = var.name
  environment         = var.environment
  resource_group_name = module.network.resource_group_name
  location            = module.network.resource_group_location
  kubernetes_version  = var.kubernetes_version
  subnet_id           = module.network.subnet_id
  identity_id         = module.iam.identity_id
  service_cidr        = var.service_cidr
  dns_service_ip      = var.dns_service_ip
  system_node_count   = var.system_node_count
  system_vm_size      = var.system_vm_size
  availability_zones  = var.availability_zones
}
