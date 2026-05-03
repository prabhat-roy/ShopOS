module "network" {
  source = "./modules/network"

  name        = var.name
  region      = var.region
  subnet_cidr = var.subnet_cidr
}
