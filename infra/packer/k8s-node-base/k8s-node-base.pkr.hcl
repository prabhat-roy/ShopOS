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
  timestamp  = regex_replace(formatdate("YYYY-MM-DD-hhmm", timestamp()), "[:]", "-")
  image_name = "shopos-k8s-node-${local.timestamp}"
}

variable "gcp_project_id" {
  type    = string
  default = "shopos-prod"
}

# ─── AWS — base image for self-managed K8s worker nodes ──────────────────────

source "amazon-ebs" "k8s-node" {
  ami_name      = local.image_name
  instance_type = "m5.large"
  region        = "us-east-1"

  source_ami_filter {
    filters = {
      name                = "ubuntu/images/hvm-ssd-gp3/ubuntu-*-26.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    owners      = ["099720109477"] # Canonical
    most_recent = true
  }

  ssh_username = "ubuntu"

  launch_block_device_mappings {
    device_name           = "/dev/sda1"
    volume_size           = 50
    volume_type           = "gp3"
    delete_on_termination = true
  }

  tags = {
    Name      = local.image_name
    Project   = "shopos"
    Role      = "k8s-node"
    BaseOS    = "ubuntu-26.04"
    Built     = local.timestamp
    ManagedBy = "packer"
  }
}

# ─── GCE — base image for GKE custom node pools ──────────────────────────────

source "googlecompute" "k8s-node" {
  project_id          = var.gcp_project_id
  source_image_family = "ubuntu-2604-lts-amd64"
  zone                = "us-central1-a"
  machine_type        = "n2-standard-2"
  disk_size           = 50
  image_name          = local.image_name
  image_family        = "shopos-k8s-node"
  image_labels = {
    project = "shopos"
    role    = "k8s-node"
    managed = "packer"
  }
  ssh_username = "ubuntu"
}

# ─── Build ────────────────────────────────────────────────────────────────────

build {
  name = "k8s-node"
  sources = [
    "source.amazon-ebs.k8s-node",
    "source.googlecompute.k8s-node"
  ]

  # Reuse Ansible role for K8s node bootstrap
  provisioner "ansible" {
    playbook_file = "../../ansible/playbooks/k8s-node.yml"
    extra_arguments = [
      "--extra-vars", "packer_build=true"
    ]
    ansible_env_vars = [
      "ANSIBLE_HOST_KEY_CHECKING=False",
      "ANSIBLE_NOCOLOR=True"
    ]
  }

  # K8s runtime + system hardening
  provisioner "shell" {
    inline = [
      # Container runtime — containerd (CRI-compatible)
      "sudo apt-get update",
      "sudo apt-get install -y containerd runc",
      "sudo mkdir -p /etc/containerd",
      "sudo containerd config default | sudo tee /etc/containerd/config.toml >/dev/null",
      "sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml",
      "sudo systemctl enable containerd",

      # Kernel modules + sysctl for K8s networking
      "echo 'overlay' | sudo tee -a /etc/modules-load.d/k8s.conf",
      "echo 'br_netfilter' | sudo tee -a /etc/modules-load.d/k8s.conf",
      "sudo modprobe overlay && sudo modprobe br_netfilter",
      "echo 'net.bridge.bridge-nf-call-iptables  = 1' | sudo tee -a /etc/sysctl.d/k8s.conf",
      "echo 'net.bridge.bridge-nf-call-ip6tables = 1' | sudo tee -a /etc/sysctl.d/k8s.conf",
      "echo 'net.ipv4.ip_forward                 = 1' | sudo tee -a /etc/sysctl.d/k8s.conf",
      "sudo sysctl --system",

      # Disable swap (required by kubelet)
      "sudo swapoff -a && sudo sed -i '/swap/d' /etc/fstab",

      # kubeadm / kubelet / kubectl pinned to 1.31
      "sudo apt-get install -y apt-transport-https ca-certificates curl gpg",
      "curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/k8s.gpg",
      "echo 'deb [signed-by=/etc/apt/keyrings/k8s.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/k8s.list",
      "sudo apt-get update",
      "sudo apt-get install -y kubelet=1.31.* kubeadm=1.31.* kubectl=1.31.*",
      "sudo apt-mark hold kubelet kubeadm kubectl",
      "sudo systemctl enable kubelet",

      # CIS hardening basics
      "echo 'kernel.kptr_restrict       = 2'                | sudo tee -a /etc/sysctl.d/99-cis.conf",
      "echo 'kernel.dmesg_restrict      = 1'                | sudo tee -a /etc/sysctl.d/99-cis.conf",
      "echo 'fs.protected_hardlinks     = 1'                | sudo tee -a /etc/sysctl.d/99-cis.conf",
      "echo 'fs.protected_symlinks      = 1'                | sudo tee -a /etc/sysctl.d/99-cis.conf",

      # Pre-pull common DaemonSet images to speed first-boot
      "sudo crictl pull docker.io/calico/cni:latest 2>/dev/null || true",
      "sudo crictl pull docker.io/calico/node:latest 2>/dev/null || true",
      "sudo crictl pull quay.io/cilium/cilium:stable 2>/dev/null || true",

      # Cleanup
      "sudo apt-get clean && sudo rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*"
    ]
  }

  # Verify
  provisioner "shell" {
    inline = [
      "containerd --version",
      "kubelet --version",
      "kubeadm version",
      "kubectl version --client",
      "echo 'K8s node base image verified'"
    ]
  }

  post-processor "manifest" {
    output     = "manifest.json"
    strip_path = true
  }
}
