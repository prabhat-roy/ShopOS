terraform {
  backend "s3" {
    # bucket, key, region, dynamodb_table are passed via -backend-config in k8s-tf-init.groovy
    # bucket         = "shopos-tfstate-<account_id>"
    # key            = "eks/terraform.tfstate"
    # region         = "us-east-1"
    # dynamodb_table = "shopos-tfstate-lock"
    # encrypt        = true
  }
}
