module "vm" {
  source = "./modules/vm"

  name                  = var.name
  environment           = var.environment
  region                = var.region
  zone                  = var.zone
  subnet_id             = module.network.subnet_id
  service_account_email = module.iam.service_account_email
  vm_size               = var.vm_size
  disk_size_gb          = var.disk_size_gb
  admin_username        = var.admin_username
  ssh_pub_key           = local.ssh_pub_key
  startup_script        = file("${path.module}/../../../../scripts/bash/jenkins-install.sh")
}
