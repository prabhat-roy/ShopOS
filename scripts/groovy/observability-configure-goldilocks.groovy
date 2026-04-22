def call() {
    sh '''
        echo "=== Configure Goldilocks ==="
        kubectl rollout status deployment/goldilocks -n monitoring --timeout=60s || true
        # Label all namespaces for VPA recommendations
        for ns in kafka rabbitmq nats prometheus grafana; do
            kubectl label namespace $ns goldilocks.fairwinds.com/enabled=true --overwrite 2>/dev/null || true
        done
        echo "Goldilocks enabled on core namespaces"
    '''
}
return this
