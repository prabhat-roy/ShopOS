terraform {
  backend "gcs" {
    # bucket and prefix are passed via -backend-config in k8s-tf-init.groovy
    # bucket = "shopos-tfstate-<project_id>"
    # prefix = "gke"
  }
}
