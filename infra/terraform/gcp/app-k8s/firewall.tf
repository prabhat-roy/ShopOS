module "firewall" {
  source = "./modules/firewall"

  name         = var.name
  network_name = module.network.network_name
  master_cidr  = var.master_cidr
}
