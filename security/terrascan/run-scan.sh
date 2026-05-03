#!/usr/bin/env bash
# =============================================================================
# ShopOS — Terrascan IaC Security Scan
#
# Scans Terraform, OpenTofu, Helm charts, Crossplane, and raw Kubernetes
# manifests for security misconfigurations.
#
# Usage:
#   ./security/terrascan/run-scan.sh
#   ./security/terrascan/run-scan.sh --severity HIGH
#   ./security/terrascan/run-scan.sh --format json
#   ./security/terrascan/run-scan.sh --fail-on-violations
#
# Output: SARIF files written to security/terrascan/results/
# =============================================================================

set -euo pipefail

# ── Parse flags ───────────────────────────────────────────────────────────────
SEVERITY="MEDIUM"
FORMAT="sarif"
FAIL_ON_VIOLATIONS=false
OUTPUT_DIR="security/terrascan/results"

for arg in "$@"; do
    case $arg in
        --severity=*)        SEVERITY="${arg#*=}" ;;
        --format=*)          FORMAT="${arg#*=}" ;;
        --fail-on-violations) FAIL_ON_VIOLATIONS=true ;;
        *) echo "Unknown option: $arg" && exit 1 ;;
    esac
done

# ── Resolve repo root ──────────────────────────────────────────────────────────
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$REPO_ROOT"

# ── Validate terrascan is installed ───────────────────────────────────────────
if ! command -v terrascan &>/dev/null; then
    echo "ERROR: terrascan is not installed."
    echo "  curl -L 'https://github.com/tenable/terrascan/releases/latest/download/terrascan_Linux_x86_64.tar.gz' | tar xz"
    echo "  sudo mv terrascan /usr/local/bin/"
    exit 1
fi

mkdir -p "$OUTPUT_DIR"

OVERALL_EXIT=0

# ── Helper: run a scan target ─────────────────────────────────────────────────
run_scan() {
    local name="$1"
    local iac_type="$2"
    local path="$3"
    local output_file="${OUTPUT_DIR}/${name}.${FORMAT}"

    if [[ ! -d "$path" ]]; then
        echo "  [SKIP] $path does not exist — skipping $name"
        return 0
    fi

    echo ""
    echo "==> Scanning: $name ($path) [iac-type=$iac_type]"

    terrascan scan \
        --iac-type "$iac_type" \
        --iac-dir "$path" \
        --severity "$SEVERITY" \
        --config-path "security/terrascan/config.toml" \
        --output "$FORMAT" \
        --log-level "error" \
        2>&1 | tee "$output_file"

    local exit_code=${PIPESTATUS[0]}

    # terrascan exits 3 when violations found (not an error, just findings)
    if [[ $exit_code -eq 3 ]]; then
        echo "  [VIOLATIONS FOUND] $name — results in $output_file"
        if [[ "${FAIL_ON_VIOLATIONS}" == "true" ]]; then
            OVERALL_EXIT=1
        fi
    elif [[ $exit_code -ne 0 ]]; then
        echo "  [ERROR] Scan failed for $name (exit $exit_code)"
        OVERALL_EXIT=1
    else
        echo "  [CLEAN] $name — no violations at severity $SEVERITY+"
    fi
}

echo "=================================================="
echo "  ShopOS Terrascan Security Scan"
echo "  Severity : ${SEVERITY}+"
echo "  Format   : ${FORMAT}"
echo "  Output   : ${OUTPUT_DIR}/"
echo "=================================================="

# ── Terraform scans ───────────────────────────────────────────────────────────
run_scan "terraform-aws-eks"       "terraform" "infra/terraform/eks"
run_scan "terraform-gcp-gke"       "terraform" "infra/terraform/gke"
run_scan "terraform-azure-aks"     "terraform" "infra/terraform/aks"
run_scan "terraform-jenkins"       "terraform" "infra/terraform/jenkins"

# ── OpenTofu scans ────────────────────────────────────────────────────────────
run_scan "opentofu-aws"            "terraform" "infra/opentofu/aws/app-k8s"
run_scan "opentofu-gcp"            "terraform" "infra/opentofu/gcp/app-k8s"
run_scan "opentofu-azure"          "terraform" "infra/opentofu/azure/app-k8s"

# ── Helm chart scans ──────────────────────────────────────────────────────────
run_scan "helm-service-charts"     "helm"      "helm/charts"
run_scan "helm-gitops-charts"      "helm"      "gitops/charts"
run_scan "helm-registry-charts"    "helm"      "registry/charts"

# ── Raw Kubernetes manifest scans ─────────────────────────────────────────────
run_scan "k8s-namespaces"          "k8s"       "kubernetes/namespaces"
run_scan "k8s-rbac"                "k8s"       "kubernetes/rbac"
run_scan "k8s-network-policies"    "k8s"       "kubernetes/network-policies"
run_scan "k8s-resource-quotas"     "k8s"       "kubernetes/resource-quotas"

# ── Summary ────────────────────────────────────────────────────────────────────
echo ""
echo "=================================================="
echo "  Scan complete. Results saved to: ${OUTPUT_DIR}/"
ls -lh "${OUTPUT_DIR}/"
echo ""

if [[ "${OVERALL_EXIT}" -ne 0 ]]; then
    echo "  RESULT: Violations found or scan errors occurred."
    echo "  Review the SARIF files and address HIGH/CRITICAL findings."
    exit 1
else
    echo "  RESULT: All scans passed (no violations at ${SEVERITY}+ severity)."
fi
echo "=================================================="
