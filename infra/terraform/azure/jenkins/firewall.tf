module "firewall" {
  source = "./modules/firewall"

  name                = var.name
  environment         = var.environment
  resource_group_name = module.network.resource_group_name
  location            = module.network.resource_group_location
  subnet_id           = module.network.subnet_id
  ssh_source_cidr     = local.caller_cidr
  ui_source_cidr      = var.ui_source_cidr
}
