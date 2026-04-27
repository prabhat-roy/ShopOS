# OpenTofu â€” GCP, Azure, and Non-Production Environments

Scope: GCP production, Azure production, and ALL non-production environments (dev/staging across all clouds).

OpenTofu is the open-source Terraform fork (Linux Foundation). It manages:
1. GCP GKE production cluster and GCP-native resources
2. Azure AKS production cluster and Azure-native resources
3. Dev and staging EKS/GKE/AKS clusters across all clouds
4. Any infrastructure that must remain fully open-source (no HashiCorp BSL)

## Why OpenTofu (not Terraform) Here?

- No HashiCorp BSL licence concern â€” fully OSI-approved open source
- Feature-compatible with Terraform (same provider registry)
- Used for non-production to keep provider/module parity with Terraform production modules
- GCP and Azure teams prefer OpenTofu for cloud-agnostic portability

## Directory Structure

```
opentofu/
â”œâ”€â”€ aws/
â”‚   â”œâ”€â”€ dev/            â† Dev EKS cluster (t3.medium, single-AZ, no HA)
â”‚   â”‚   â”œâ”€â”€ main.tofu
â”‚   â”‚   â””â”€â”€ variables.tofu
â”‚   â””â”€â”€ staging/        â† Staging EKS cluster (m5.large, multi-AZ, mirrors prod)
â”‚       â”œâ”€â”€ main.tofu
â”‚       â””â”€â”€ variables.tofu
â”‚
â”œâ”€â”€ gcp/
â”‚   â”œâ”€â”€ gke/            â† GKE Autopilot cluster (production)
â”‚   â”‚   â”œâ”€â”€ main.tofu   â† GKE cluster, node pools, Workload Identity
â”‚   â”‚   â”œâ”€â”€ network.tofu â† VPC, subnets, Cloud NAT, firewall rules
â”‚   â”‚   â”œâ”€â”€ cloudsql.tofu â† Cloud SQL PostgreSQL (analytics domain)
â”‚   â”‚   â”œâ”€â”€ bigquery.tofu â† BigQuery datasets for analytics-ai domain
â”‚   â”‚   â”œâ”€â”€ pubsub.tofu â† Pub/Sub topics for GCP-native event bus
â”‚   â”‚   â”œâ”€â”€ artifact_registry.tofu â† Container image registry
â”‚   â”‚   â””â”€â”€ cloud_run.tofu â† Cloud Run services (ML model serving)
â”‚   â”œâ”€â”€ dev/            â† GCP dev environment
â”‚   â””â”€â”€ staging/        â† GCP staging environment
â”‚
â””â”€â”€ azure/
    â”œâ”€â”€ aks/            â† AKS cluster (production)
    â”‚   â”œâ”€â”€ main.tofu   â† AKS cluster, node pools, Managed Identity
    â”‚   â”œâ”€â”€ network.tofu â† VNet, subnets, NSG, Azure Firewall
    â”‚   â”œâ”€â”€ acr.tofu    â† Azure Container Registry
    â”‚   â”œâ”€â”€ keyvault.tofu â† Azure Key Vault (C# service secrets)
    â”‚   â”œâ”€â”€ cosmosdb.tofu â† Cosmos DB (cart-service, return-refund-service)
    â”‚   â””â”€â”€ servicebus.tofu â† Azure Service Bus (RabbitMQ alternative for Azure)
    â”œâ”€â”€ dev/
    â””â”€â”€ staging/
```

## What OpenTofu Does NOT Manage

| Resource Type | Managed By |
|---|---|
| AWS production EKS | Terraform (`infra/terraform/aws/`) |
| K8s workloads | ArgoCD / Flux (GitOps) |
| K8s-native infra | Crossplane (`infra/crossplane/`) |
| Server OS config | Ansible (`infra/ansible/`) |
| VM/container base images | Packer (`infra/packer/`) |

## Key Differences vs Terraform Directory

| Aspect | Terraform | OpenTofu |
|---|---|---|
| Clouds | AWS only | GCP + Azure + AWS dev/staging |
| Environments | Production only | Dev + Staging + GCP/Azure prod |
| Licence | HashiCorp BSL | OSI-approved open source |
| State backend | S3 (AWS) | GCS (GCP), Azure Blob (Azure), S3 (AWS dev) |
| CI integration | Atlantis | Scalr (open-source alternative) |

## Usage

```bash
# GCP production
cd infra/opentofu/gcp/gke
tofu init -backend-config=../../backend-gcs.hcl
tofu plan -var-file=../../gcp-prod.tfvars

# AWS dev environment
cd infra/opentofu/aws/dev
tofu init
tofu plan -var-file=dev.tfvars
tofu apply -auto-approve   # Dev only â€” auto-approve allowed

# Azure production
cd infra/opentofu/azure/aks
tofu init -backend-config=../../backend-azurerm.hcl
tofu plan -var-file=../../azure-prod.tfvars
```
