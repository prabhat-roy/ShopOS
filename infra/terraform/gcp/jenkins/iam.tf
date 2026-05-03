module "iam" {
  source = "./modules/iam"

  name       = var.name
  project_id = var.project_id
}
