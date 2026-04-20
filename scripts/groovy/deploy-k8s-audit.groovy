def call() {
    sh 'mkdir -p reports/k8s-audit'

    // ── kube-bench — CIS Kubernetes Benchmark ────────────────────────────────
    sh """
        echo "=== K8s Audit: kube-bench (CIS) ==="
        docker run --rm \
            -v /var/run/docker.sock:/var/run/docker.sock \
            -v \${KUBECONFIG}:/root/.kube/config:ro \
            aquasec/kube-bench:latest \
            --json \
            > reports/k8s-audit/kube-bench.json 2>&1 || true
    """

    // ── kube-hunter — K8s penetration test ───────────────────────────────────
    sh """
        echo "=== K8s Audit: kube-hunter ==="
        docker run --rm \
            -v \${KUBECONFIG}:/root/.kube/config:ro \
            aquasec/kube-hunter:latest \
            --remote \$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' | sed 's|https://||') \
            --report json \
            > reports/k8s-audit/kube-hunter.json 2>&1 || true
    """

    // ── Kubescape — NSA/MITRE ATT&CK framework compliance ────────────────────
    sh """
        echo "=== K8s Audit: Kubescape ==="
        docker run --rm \
            -v \${KUBECONFIG}:/root/.kube/config:ro \
            -v \${WORKSPACE}/kubernetes:/manifests:ro \
            kubescape/kubescape:latest scan \
            framework NSA,MITRE \
            --format json \
            --output /tmp/kubescape.json 2>/dev/null || true
        cp /tmp/kubescape.json reports/k8s-audit/kubescape.json 2>/dev/null || true
    """

    // ── Kubeaudit — manifest + live cluster audit ─────────────────────────────
    sh """
        echo "=== K8s Audit: Kubeaudit (live cluster) ==="
        docker run --rm \
            -v \${KUBECONFIG}:/root/.kube/config:ro \
            shopify/kubeaudit:latest all \
            --format json \
            > reports/k8s-audit/kubeaudit-cluster.json 2>&1 || true
    """

    // ── cnspec — cloud-native security posture ────────────────────────────────
    sh """
        echo "=== K8s Audit: cnspec ==="
        docker run --rm \
            -v \${KUBECONFIG}:/root/.kube/config:ro \
            mondoo/cnspec:latest scan k8s \
            --output json \
            > reports/k8s-audit/cnspec.json 2>&1 || true
    """

    echo 'Kubernetes security audit complete — reports/k8s-audit/'
}
return this
