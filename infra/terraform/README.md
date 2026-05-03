# Terraform — All Cloud Infrastructure (AWS + GCP + Azure in Parallel)

All cloud infrastructure for ShopOS. AWS, GCP, and Azure each get their own folder, and
inside each cloud there's one folder per workload. Every workload directory is **fully
self-contained** — its own providers, variables, state, and a private `modules/` folder
that ships only the modules that workload uses.

## Who runs what

| Workload | Run from | Why |
|---|---|---|
| `<cloud>/jenkins/` | **Laptop** (`terraform apply`) | Chicken-and-egg: Jenkins doesn't exist yet. This is the only IaC step performed by hand. |
| `<cloud>/app-k8s/` | **Jenkins pipeline** ([k8s-infra.Jenkinsfile](../../ci/jenkins/k8s-infra.Jenkinsfile)) | Once Jenkins is up, all subsequent infra is operated through it. Param `IaC_TOOL=terraform` for prod, `IaC_TOOL=opentofu` for dev/staging — both consume the same module shape. |

Quality / cost / drift jobs ([infra-quality.Jenkinsfile](../../ci/jenkins/infra-quality.Jenkinsfile))
also run from Jenkins on a weekly cron (Infracost, Driftctl, tflint, Atlantis-validate, inframap).

## Layout

```
infra/terraform/
├── aws/
│   ├── jenkins/        ← Jenkins server (Ubuntu 26.04 + EC2 + EIP + IAM + SG)
│   │   ├── modules/{network, firewall, iam, vm}/
│   │   ├── providers.tf, variables.tf, data_caller_ip.tf
│   │   ├── network.tf, firewall.tf, iam.tf, vm.tf      ← module callers
│   │   ├── outputs.tf, terraform.tfvars.example
│   └── app-k8s/        ← EKS Auto Mode cluster (VPC + subnets + NAT + IAM + EKS)
│       ├── modules/{network, firewall, iam, eks}/
│       └── providers.tf, network.tf, firewall.tf, iam.tf, eks.tf, outputs.tf, ...
├── gcp/
│   ├── jenkins/        ← Compute Engine + IAM SA + firewall
│   │   ├── modules/{network, firewall, iam, vm}/
│   │   └── ...
│   └── app-k8s/        ← GKE Autopilot (VPC + subnet + Cloud NAT + GKE)
│       ├── modules/{network, firewall, gke}/
│       └── ...
└── azure/
    ├── jenkins/        ← VM + NIC + NSG + identity
    │   ├── modules/{network, firewall, iam, vm}/
    │   └── ...
    └── app-k8s/        ← AKS (RG + VNet + identity + AKS)
        ├── modules/{network, iam, aks}/
        └── ...
```

Each workload folder is **independent**:

- Its own `providers.tf`, `variables.tf`, `outputs.tf`, `terraform.tfvars.example`.
- Its own `modules/` directory holding only the modules it consumes.
- Its own state (no shared state between workloads or clouds).
- One `*.tf` at the top level per module instantiation (e.g., `network.tf` calls
  `./modules/network`, `vm.tf` calls `./modules/vm`).

The `app-k8s` naming leaves room for additional cluster types later
(`infra-k8s`, `analytics-k8s`, etc.) without renaming.

## Convention: one resource per .tf file (inside each module)

Every resource lives in its own `.tf` file, named after the resource. Examples from
`aws/jenkins/modules/`:

| Module | File | Resource |
|---|---|---|
| `network/` | `vpc.tf` | `aws_vpc.this` |
| `network/` | `subnet.tf` | `aws_subnet.this` |
| `network/` | `internet_gateway.tf` | `aws_internet_gateway.this` |
| `network/` | `route_table.tf` | `aws_route_table.this` |
| `network/` | `route_table_association.tf` | `aws_route_table_association.this` |
| `firewall/` | `security_group.tf` | `aws_security_group.this` |
| `firewall/` | `ssh.tf` | `aws_vpc_security_group_ingress_rule.ssh` |
| `firewall/` | `ui.tf` | `aws_vpc_security_group_ingress_rule.ui` |
| `firewall/` | `egress.tf` | `aws_vpc_security_group_egress_rule.all` |
| `iam/` | `iam_role.tf` | `aws_iam_role.this` |
| `iam/` | `iam_role_policy_attachment.tf` | `aws_iam_role_policy_attachment.admin` |
| `iam/` | `iam_instance_profile.tf` | `aws_iam_instance_profile.this` |
| `vm/` | `data_ami.tf` | `data.aws_ami.ubuntu` |
| `vm/` | `instance.tf` | `aws_instance.this` |
| `vm/` | `eip.tf` | `aws_eip.this` |

Each module also has its own `variables.tf` and `outputs.tf`. The top-level
`data_caller_ip.tf` holds the `data.http.caller_ip` lookup and the SSH-key
auto-detect `locals` (GCP/Azure only).

## Bootstrap mechanism: cloud-init / user_data

The Jenkins install script (`scripts/bash/jenkins-install.sh`) is fed to the VM via
**cloud-init**, the same on all three clouds:

| Cloud | Mechanism | Resource attribute |
|---|---|---|
| AWS | EC2 `user_data` | `aws_instance.this.user_data` |
| GCP | Compute `metadata_startup_script` | `google_compute_instance.this.metadata_startup_script` |
| Azure | VM `custom_data` (base64-encoded) | `azurerm_linux_virtual_machine.this.custom_data` |

No `null_resource` provisioner. No `remote-exec`. No SSH from the apply host required.
`terraform apply` returns once the VM is created. Bootstrap progress streams to
`/var/log/cloud-init-output.log` on the VM. Jenkins is reachable on `:8080` once
`/var/lib/jenkins/jenkins-setup-complete` exists.

## SSH

| Cloud | Connect with | Key handling |
|---|---|---|
| AWS | `ssh -i ~/.ssh/<key>.pem ubuntu@<ip>` | Pre-existing AWS EC2 key pair (`var.key_name`) |
| GCP | `ssh ubuntu@<ip>` (no `-i` flag) | Auto-detects `~/.ssh/id_ed25519.pub` → `id_rsa.pub`; injected into VM metadata |
| Azure | `ssh ubuntu@<ip>` (no `-i` flag) | Same auto-detect as GCP |

## OS

All three clouds run **Ubuntu 26.04 LTS**:

- AWS AMI filter: `ubuntu/images/hvm-ssd-gp3/ubuntu-*-26.04-amd64-server-*` (`most_recent = true`)
- GCP image: `ubuntu-os-cloud/ubuntu-2604-lts-amd64`
- Azure SKU: `Canonical / ubuntu-26_04-lts / server`

## Usage

Same shape across every cloud + workload:

```bash
cd infra/terraform/<aws|gcp|azure>/<jenkins|app-k8s>
cp terraform.tfvars.example terraform.tfvars   # edit values
terraform init
terraform apply
terraform output                                # jenkins_url / kubeconfig_command / ssh_command
```

Examples:

```bash
# Jenkins on AWS
cd infra/terraform/aws/jenkins
# Edit terraform.tfvars: set key_name to an existing AWS key pair name
terraform init && terraform apply
ssh -i ~/.ssh/us-east-1.pem ubuntu@$(terraform output -raw jenkins_public_ip)

# Jenkins on GCP
cd infra/terraform/gcp/jenkins
# Edit terraform.tfvars: set project_id; SSH key auto-detected
terraform init && terraform apply
ssh ubuntu@$(terraform output -raw jenkins_public_ip)

# Jenkins on Azure
cd infra/terraform/azure/jenkins
# Edit terraform.tfvars: set subscription_id; SSH key auto-detected
terraform init && terraform apply
ssh ubuntu@$(terraform output -raw jenkins_public_ip)

# K8s cluster on AWS
cd infra/terraform/aws/app-k8s
terraform init && terraform apply
$(terraform output -raw kubeconfig_command)
kubectl get nodes
```

## Atlantis GitOps

```
Developer opens PR → Atlantis runs terraform plan → Posts plan as PR comment
Tech lead approves PR → Atlantis runs terraform apply → Updates state in S3
Driftctl runs weekly → Reports drift between state and actual AWS resources
Infracost runs on PR → Reports cost delta (must be < $500/month per PR)
```
