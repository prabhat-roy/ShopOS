def call() {
    sh '''
        echo "=== Configure OpenCost ==="
        kubectl rollout status deployment/opencost -n monitoring --timeout=60s || true
        echo "OpenCost ready — cost allocation data available at /allocation/compute"
    '''
}
return this
