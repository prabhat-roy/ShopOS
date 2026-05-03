#!/usr/bin/env bash
# Trivy + Grype + Syft + Checkov consolidated scan — used in pre-commit + CI.
# Replaces the Wiz commercial product with open-source equivalents.
set -euo pipefail
TARGET="${1:-.}"

echo ">>> Trivy filesystem scan"
trivy fs --severity HIGH,CRITICAL --exit-code 1 "${TARGET}"

echo ">>> Trivy IaC scan"
trivy config --severity HIGH,CRITICAL --exit-code 0 "${TARGET}"

echo ">>> Grype scan (catches what Trivy misses)"
grype dir:"${TARGET}" --fail-on high

echo ">>> Syft SBOM (CycloneDX)"
syft "${TARGET}" -o cyclonedx-json=/tmp/sbom.cdx.json

echo ">>> Checkov (Terraform, Helm, Dockerfile)"
checkov -d "${TARGET}" --framework terraform helm dockerfile --quiet

echo ">>> Gitleaks (secrets in working tree)"
gitleaks detect --source "${TARGET}" --no-banner

echo ">>> Done"
