module "network" {
  source = "./modules/network"

  name        = var.name
  environment = var.environment
  region      = var.region
  vpc_cidr    = var.vpc_cidr
  subnet_cidr = var.subnet_cidr
}
