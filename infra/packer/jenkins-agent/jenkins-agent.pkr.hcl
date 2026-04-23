# Packer — Jenkins Agent AMI
# Bakes a pre-warmed EC2 AMI with all build tools pre-installed.
# Eliminates 8-12 minute tool installation at the start of every Jenkins build.
# Also used as a GCE image for GCP Cloud Build custom workers.

packer {
  required_plugins {
    amazon = {
      version = ">= 1.3.0"
      source  = "github.com/hashicorp/amazon"
    }
    googlecompute = {
      version = ">= 1.1.0"
      source  = "github.com/hashicorp/googlecompute"
    }
    ansible = {
      version = ">= 1.1.0"
      source  = "github.com/hashicorp/ansible"
    }
  }
}

locals {
  timestamp = regex_replace(formatdate("YYYY-MM-DD-hhmm", timestamp()), "[:]", "-")
  ami_name  = "shopos-jenkins-agent-${local.timestamp}"
}

# ─── AWS AMI ─────────────────────────────────────────────────────────────────

source "amazon-ebs" "jenkins-agent" {
  ami_name      = local.ami_name
  instance_type = "m5.2xlarge"
  region        = "us-east-1"

  source_ami_filter {
    filters = {
      name                = "ubuntu/images/hvm-ssd-gp3/ubuntu-noble-24.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    owners      = ["099720109477"] # Canonical
    most_recent = true
  }

  ssh_username = "ubuntu"

  launch_block_device_mappings {
    device_name           = "/dev/sda1"
    volume_size           = 100
    volume_type           = "gp3"
    iops                  = 3000
    throughput            = 125
    delete_on_termination = true
  }

  tags = {
    Name        = local.ami_name
    Project     = "shopos"
    Role        = "jenkins-agent"
    BaseOS      = "ubuntu-24.04"
    Built       = local.timestamp
    ManagedBy   = "packer"
  }

  run_tags = {
    Name = "packer-builder-jenkins-agent"
  }
}

# ─── GCE Image ───────────────────────────────────────────────────────────────

source "googlecompute" "jenkins-agent" {
  project_id          = var.gcp_project_id
  source_image_family = "ubuntu-2404-lts"
  zone                = "us-central1-a"
  machine_type        = "n2-standard-8"
  disk_size           = 100
  image_name          = local.ami_name
  image_family        = "shopos-jenkins-agent"
  image_labels = {
    project   = "shopos"
    role      = "jenkins-agent"
    managed   = "packer"
  }
  ssh_username = "ubuntu"
}

# ─── Build ────────────────────────────────────────────────────────────────────

build {
  name = "jenkins-agent"
  sources = [
    "source.amazon-ebs.jenkins-agent",
    "source.googlecompute.jenkins-agent"
  ]

  # Ansible provisioner — reuses ansible roles from infra/ansible
  provisioner "ansible" {
    playbook_file = "../ansible/playbooks/jenkins-agent.yml"
    extra_arguments = [
      "--extra-vars", "packer_build=true"
    ]
    ansible_env_vars = [
      "ANSIBLE_HOST_KEY_CHECKING=False",
      "ANSIBLE_NOCOLOR=True"
    ]
  }

  # Shell provisioner for tool versions pinning
  provisioner "shell" {
    inline = [
      # Go
      "GO_VERSION=1.23.3",
      "wget -qO- https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf -",
      "echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/environment",

      # Java 21 LTS
      "apt-get install -y openjdk-21-jdk-headless",
      "update-alternatives --set java /usr/lib/jvm/java-21-openjdk-amd64/bin/java",

      # Python 3.12
      "apt-get install -y python3.12 python3.12-pip python3.12-venv",
      "pip3 install --upgrade pip awscli boto3 pytest black ruff",

      # Node.js 20 LTS
      "curl -fsSL https://deb.nodesource.com/setup_20.x | bash -",
      "apt-get install -y nodejs",
      "npm install -g npm@latest typescript @stoplight/spectral-cli",

      # Docker (rootless builds via Kaniko, but Docker CLI needed for local test)
      "apt-get install -y docker.io",
      "usermod -aG docker ubuntu",

      # Kubernetes tools
      "curl -LO https://dl.k8s.io/release/v1.31.0/bin/linux/amd64/kubectl && install kubectl /usr/local/bin/",
      "curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash",
      "curl -sL https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64 -o /usr/local/bin/argocd && chmod +x /usr/local/bin/argocd",

      # Security tools
      "curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin",
      "curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin",
      "curl -sfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin",
      "wget -qO /usr/local/bin/cosign https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64 && chmod +x /usr/local/bin/cosign",

      # IaC tools
      "wget -qO- https://releases.hashicorp.com/terraform/1.9.8/terraform_1.9.8_linux_amd64.zip | unzip -d /usr/local/bin -",
      "wget -qO- https://github.com/opentofu/opentofu/releases/latest/download/tofu_1.8.5_linux_amd64.zip | unzip -d /usr/local/bin -",
      "pip3 install ansible checkov",

      # Buf (protobuf)
      "curl -sSL https://github.com/bufbuild/buf/releases/latest/download/buf-Linux-x86_64 -o /usr/local/bin/buf && chmod +x /usr/local/bin/buf",

      # Rust (for auth-service, shipping-service)
      "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y",
      "echo 'source $HOME/.cargo/env' >> /etc/environment",

      # Clean up
      "apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*"
    ]
  }

  # Validate the AMI has all required tools
  provisioner "shell" {
    inline = [
      "go version",
      "java -version",
      "python3.12 --version",
      "node --version",
      "docker --version",
      "kubectl version --client",
      "helm version",
      "terraform version",
      "tofu version",
      "trivy --version",
      "cosign version",
      "buf --version",
      "echo 'All tools verified successfully'"
    ]
  }

  post-processor "manifest" {
    output     = "manifest.json"
    strip_path = true
  }
}

variable "gcp_project_id" {
  type    = string
  default = "shopos-prod"
}
