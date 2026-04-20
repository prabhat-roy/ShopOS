# Ansible — ShopOS

Node bootstrapping and configuration management for Jenkins servers and Kubernetes nodes across AWS, GCP, and Azure.

## Structure

```
ansible/
├── ansible.cfg           ← Ansible configuration (remote_user, SSH, forks)
├── requirements.yml      ← Ansible Galaxy collection dependencies
├── inventory/
│   ├── hosts.ini         ← Static inventory for manual/local runs
│   ├── aws.yml           ← Dynamic inventory via amazon.aws.aws_ec2
│   ├── gcp.yml           ← Dynamic inventory via google.cloud.gcp_compute
│   └── azure.yml         ← Dynamic inventory via azure.azcollection.azure_rm
├── group_vars/
│   ├── all.yml           ← Variables for every host (packages, SSH, timezone)
│   ├── jenkins.yml       ← Jenkins-specific vars (plugins, tool versions)
│   └── k8s_nodes.yml     ← K8s node vars (version, kernel modules, sysctl)
├── playbooks/
│   ├── common.yml        ← Baseline for all hosts (packages, SSH hardening, ufw)
│   ├── jenkins.yml       ← Full Jenkins server setup (common + docker + tools + jenkins)
│   └── k8s-node.yml      ← Kubernetes node bootstrap (common + containerd + kubelet)
└── roles/
    ├── common/           ← apt update, packages, SSH hardening, ufw, swap off
    ├── docker/           ← Docker CE install, socket permissions, group membership
    ├── jenkins/          ← Jenkins install, wizard disable, admin user, plugins
    ├── tools/            ← Go, Node.js, Terraform, kubectl, Helm, Maven, Rust, .NET
    └── k8s-node/         ← containerd, kubelet, kubeadm, kubectl, kernel settings
```

## Quick Start

### 1. Install dependencies

```bash
pip3 install ansible boto3 google-auth azure-mgmt-compute
ansible-galaxy collection install -r requirements.yml
```

### 2. Set up inventory

**Static (manual):** Edit `inventory/hosts.ini` and add your server IPs.

**Dynamic AWS:**
```bash
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_DEFAULT_REGION=us-east-1
ansible-inventory -i inventory/aws.yml --list
```

**Dynamic GCP:**
```bash
export GCP_PROJECT_ID=your-project
gcloud auth application-default login
ansible-inventory -i inventory/gcp.yml --list
```

**Dynamic Azure:**
```bash
az login
ansible-inventory -i inventory/azure.yml --list
```

### 3. Run playbooks

```bash
# Apply baseline to all hosts
ansible-playbook playbooks/common.yml -i inventory/hosts.ini

# Bootstrap a Jenkins server
ansible-playbook playbooks/jenkins.yml -i inventory/hosts.ini

# Bootstrap Jenkins on AWS (tag Role=jenkins)
ansible-playbook playbooks/jenkins.yml -i inventory/aws.yml --limit role_jenkins

# Bootstrap Kubernetes nodes
ansible-playbook playbooks/k8s-node.yml -i inventory/hosts.ini

# Override Jenkins password at runtime
ansible-playbook playbooks/jenkins.yml -i inventory/hosts.ini \
  --extra-vars "jenkins_admin_password=mysecurepassword"

# Dry run (check mode)
ansible-playbook playbooks/jenkins.yml -i inventory/hosts.ini --check

# Verbose output
ansible-playbook playbooks/jenkins.yml -i inventory/hosts.ini -vv
```

## Roles

| Role | What it does |
|---|---|
| `common` | apt upgrade, common packages, SSH hardening, ufw, timezone, swap off |
| `docker` | Docker CE install, socket permissions, adds jenkins/ubuntu to docker group |
| `jenkins` | Java 21, Jenkins install, wizard disabled, admin user, plugin install, ufw rule |
| `tools` | Go, Node.js, Terraform, kubectl, Helm, Maven, Rust, .NET SDK |
| `k8s-node` | Kernel modules, sysctl, containerd, kubelet, kubeadm, kubectl |

## Relationship to Terraform

Terraform provisions the VM (EC2/GCE/AzureVM) and writes its public IP to `infra.env`. Ansible runs after Terraform to configure what's running on that VM.

```
terraform apply          → VM created, IP written to infra.env
ansible-playbook ...     → VM configured (Jenkins, tools, Docker)
```

The `infra/terraform/aws/jenkins/provisioner.tf` calls `scripts/bash/jenkins-install.sh` as a Terraform remote-exec provisioner. The Ansible playbooks are the idiomatic replacement — use whichever fits your workflow.
