module "firewall" {
  source = "./modules/firewall"

  name             = var.name
  environment      = var.environment
  vpc_id           = module.network.vpc_id
  vpc_cidr         = module.network.vpc_cidr
  k8s_cluster_name = var.name
}
