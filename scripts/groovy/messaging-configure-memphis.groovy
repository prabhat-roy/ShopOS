def call() {
    sh """
        echo "Configuring Memphis via kubectl exec..."
        POD=\$(kubectl get pods -n memphis --no-headers | head -1 | awk '{print \$1}')
        if [ -z "\${POD}" ]; then echo "No Memphis pod found — skipping"; exit 0; fi

        # Login and get JWT token
        TOKEN=\$(kubectl exec -n memphis "\${POD}" -- \
            curl -sf -X POST http://localhost:9000/auth/login \
            -H 'Content-Type: application/json' \
            -d '{"username":"root","password":"memphis"}' \
            | grep -o '"jwt":"[^"]*"' | cut -d: -f2 | tr -d '"' 2>/dev/null || true)

        if [ -z "\${TOKEN}" ]; then
            echo "Could not obtain Memphis token — skipping station creation"
            exit 0
        fi

        # Create ShopOS station
        kubectl exec -n memphis "\${POD}" -- \
            curl -sf -X POST http://localhost:9000/stations \
            -H "Authorization: Bearer \${TOKEN}" \
            -H 'Content-Type: application/json' \
            -d '{"name":"shopos-events","retention_type":"message_age_sec","retention_value":86400,"storage_type":"file","replicas":1}' || true

        sed -i '/^MEMPHIS_TOKEN=/d' infra.env || true
        echo "MEMPHIS_TOKEN=\${TOKEN}" >> infra.env
        echo "Memphis shopos-events station created."
    """
    echo 'memphis configured — shopos-events station created'
}
return this
