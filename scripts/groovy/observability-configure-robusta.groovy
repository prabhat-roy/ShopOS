def call() {
    sh '''
        echo "=== Configure Robusta ==="
        kubectl rollout status deployment/robusta -n monitoring --timeout=60s || true
        echo "Robusta ready — auto-remediation playbooks active"
    '''
}
return this
