def call() {
    sh 'mkdir -p reports/iac'

    // ── Checkov ───────────────────────────────────────────────────────────────
    sh """
        echo "=== IaC: Checkov ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            bridgecrew/checkov:latest \
            --directory /src/helm,/src/kubernetes,/src/infra/terraform \
            --output json \
            --output-file-path /src/reports/iac || true
    """

    // ── KICS ──────────────────────────────────────────────────────────────────
    sh """
        echo "=== IaC: KICS ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            checkmarx/kics:latest scan \
            -p /src/helm,/src/kubernetes \
            -o /src/reports/iac \
            --report-formats json \
            --ci || true
    """

    // ── tfsec ─────────────────────────────────────────────────────────────────
    sh """
        echo "=== IaC: tfsec ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            aquasec/tfsec:latest /src/infra/terraform \
            --format json \
            --out /src/reports/iac/tfsec.json || true
    """

    // ── Terrascan — Terraform + Helm ──────────────────────────────────────────
    sh """
        echo "=== IaC: Terrascan (Terraform) ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            tenable/terrascan:latest scan \
            -d /src/infra/terraform \
            -t terraform \
            --output json \
            > reports/iac/terrascan-terraform.json 2>&1 || true

        echo "=== IaC: Terrascan (Helm) ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            tenable/terrascan:latest scan \
            -d /src/helm \
            -t helm \
            --output json \
            > reports/iac/terrascan-helm.json 2>&1 || true
    """

    // ── Polaris — Kubernetes best practices ───────────────────────────────────
    sh """
        echo "=== IaC: Polaris ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            quay.io/fairwinds/polaris:latest audit \
            --audit-path /src/kubernetes \
            --format json \
            > reports/iac/polaris.json 2>&1 || true
    """

    // ── Kubeaudit — Kubernetes manifest security audit ────────────────────────
    sh """
        echo "=== IaC: Kubeaudit ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            shopify/kubeaudit:latest all \
            -f /src/kubernetes \
            --format json \
            > reports/iac/kubeaudit.json 2>&1 || true
    """

    // ── cnspec — cloud-native security assessment ─────────────────────────────
    sh """
        echo "=== IaC: cnspec (Mondoo) ==="
        docker run --rm \
            -v \${WORKSPACE}:/src \
            mondoo/cnspec:latest scan \
            k8s-manifest /src/kubernetes \
            --output json \
            > reports/iac/cnspec.json 2>&1 || true
    """

    echo 'IaC scanning complete — reports/iac/'
}
return this
