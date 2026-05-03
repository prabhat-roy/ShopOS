module "firewall" {
  source = "./modules/firewall"

  name            = var.name
  network_name    = module.network.network_name
  ssh_source_cidr = local.caller_cidr
  ui_source_cidr  = var.ui_source_cidr
}
