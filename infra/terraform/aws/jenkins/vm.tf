module "vm" {
  source = "./modules/vm"

  name                  = var.name
  environment           = var.environment
  subnet_id             = module.network.subnet_id
  security_group_ids    = [module.firewall.security_group_id]
  instance_profile_name = module.iam.instance_profile_name
  vm_size               = var.vm_size
  disk_size_gb          = var.disk_size_gb
  key_name              = var.key_name
  user_data             = file("${path.module}/../../../../scripts/bash/jenkins-install.sh")
}
