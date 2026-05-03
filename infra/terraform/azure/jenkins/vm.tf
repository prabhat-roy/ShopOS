module "vm" {
  source = "./modules/vm"

  name                = var.name
  environment         = var.environment
  resource_group_name = module.network.resource_group_name
  location            = module.network.resource_group_location
  subnet_id           = module.network.subnet_id
  identity_id         = module.iam.identity_id
  vm_size             = var.vm_size
  disk_size_gb        = var.disk_size_gb
  admin_username      = var.admin_username
  ssh_pub_key         = local.ssh_pub_key
  user_data           = file("${path.module}/../../../../scripts/bash/jenkins-install.sh")
}
