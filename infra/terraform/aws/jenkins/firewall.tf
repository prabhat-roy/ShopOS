module "firewall" {
  source = "./modules/firewall"

  name            = var.name
  environment     = var.environment
  vpc_id          = module.network.vpc_id
  ssh_source_cidr = local.caller_cidr
  ui_source_cidr  = var.ui_source_cidr
}
