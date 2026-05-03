module "iam" {
  source = "./modules/iam"

  name                = var.name
  environment         = var.environment
  resource_group_name = module.network.resource_group_name
  location            = module.network.resource_group_location
  subnet_id           = module.network.subnet_id
}
