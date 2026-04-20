def call() {
    sh '''
        echo "=== Configure Kong ==="

        KONG_ADMIN=$(kubectl get svc kong-admin -n kong \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "kong-admin.kong.svc.cluster.local")

        # Wait for Admin API
        for i in $(seq 1 12); do
            curl -sf "http://${KONG_ADMIN}:8001/status" >/dev/null 2>&1 && break
            echo "  Waiting for Kong Admin API... ($i/12)"
            sleep 10
        done

        # Enable rate-limiting plugin globally
        curl -sf -X POST "http://${KONG_ADMIN}:8001/plugins" \
            -d "name=rate-limiting" \
            -d "config.minute=1000" \
            -d "config.hour=10000" \
            -d "config.policy=local" 2>/dev/null || true

        # Enable request-size-limiting globally
        curl -sf -X POST "http://${KONG_ADMIN}:8001/plugins" \
            -d "name=request-size-limiting" \
            -d "config.allowed_payload_size=50" 2>/dev/null || true

        sed -i '/^KONG_ADMIN_URL=/d' infra.env
        echo "KONG_ADMIN_URL=http://${KONG_ADMIN}:8001" >> infra.env

        echo "Kong global rate-limiting and request-size plugins configured."
    '''
}
return this
