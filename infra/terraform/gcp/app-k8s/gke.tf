module "gke" {
  source = "./modules/gke"

  name               = var.name
  region             = var.region
  project_id         = var.project_id
  kubernetes_version = var.kubernetes_version
  network_name       = module.network.network_name
  subnet_name        = module.network.subnet_name
  master_cidr        = var.master_cidr
}
