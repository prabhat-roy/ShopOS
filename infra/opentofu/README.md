# OpenTofu — GCP, Azure, and Non-Production Environments

**Scope: GCP production, Azure production, and ALL non-production environments (dev/staging across all clouds).**

OpenTofu is the open-source Terraform fork (Linux Foundation). It manages:
1. GCP GKE production cluster and GCP-native resources
2. Azure AKS production cluster and Azure-native resources
3. Dev and staging EKS/GKE/AKS clusters across all clouds
4. Any infrastructure that must remain fully open-source (no HashiCorp BSL)

## Why OpenTofu (not Terraform) Here?

- No HashiCorp BSL licence concern — fully OSI-approved open source
- Feature-compatible with Terraform (same provider registry)
- Used for non-production to keep provider/module parity with Terraform production modules
- GCP and Azure teams prefer OpenTofu for cloud-agnostic portability

## Directory Structure

```
opentofu/
├── aws/
│   ├── dev/            ← Dev EKS cluster (t3.medium, single-AZ, no HA)
│   │   ├── main.tofu
│   │   └── variables.tofu
│   └── staging/        ← Staging EKS cluster (m5.large, multi-AZ, mirrors prod)
│       ├── main.tofu
│       └── variables.tofu
│
├── gcp/
│   ├── gke/            ← GKE Autopilot cluster (production)
│   │   ├── main.tofu   ← GKE cluster, node pools, Workload Identity
│   │   ├── network.tofu ← VPC, subnets, Cloud NAT, firewall rules
│   │   ├── cloudsql.tofu ← Cloud SQL PostgreSQL (analytics domain)
│   │   ├── bigquery.tofu ← BigQuery datasets for analytics-ai domain
│   │   ├── pubsub.tofu ← Pub/Sub topics for GCP-native event bus
│   │   ├── artifact_registry.tofu ← Container image registry
│   │   └── cloud_run.tofu ← Cloud Run services (ML model serving)
│   ├── dev/            ← GCP dev environment
│   └── staging/        ← GCP staging environment
│
└── azure/
    ├── aks/            ← AKS cluster (production)
    │   ├── main.tofu   ← AKS cluster, node pools, Managed Identity
    │   ├── network.tofu ← VNet, subnets, NSG, Azure Firewall
    │   ├── acr.tofu    ← Azure Container Registry
    │   ├── keyvault.tofu ← Azure Key Vault (C# service secrets)
    │   ├── cosmosdb.tofu ← Cosmos DB (cart-service, return-refund-service)
    │   └── servicebus.tofu ← Azure Service Bus (RabbitMQ alternative for Azure)
    ├── dev/
    └── staging/
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
tofu apply -auto-approve   # Dev only — auto-approve allowed

# Azure production
cd infra/opentofu/azure/aks
tofu init -backend-config=../../backend-azurerm.hcl
tofu plan -var-file=../../azure-prod.tfvars
```
