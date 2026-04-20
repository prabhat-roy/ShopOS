def call() {
    sh """
        MEMPHIS_URL=\$(grep '^MEMPHIS_HTTP_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Memphis at \${MEMPHIS_URL}..."
        until curl -sf "\${MEMPHIS_URL}/monitoring/getClusterInfo" > /dev/null 2>&1; do sleep 10; done

        # Login
        TOKEN=\$(curl -sf -X POST "\${MEMPHIS_URL}/auth/login" \
            -H "Content-Type: application/json" \
            -d '{"username":"root","password":"memphis"}' \
            | grep -o '"jwt":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create ShopOS station (topic equivalent)
        curl -sf -X POST "\${MEMPHIS_URL}/stations" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"shopos-events","retention_type":"message_age_sec","retention_value":86400,"storage_type":"file","replicas":1}' || true

        sed -i '/^MEMPHIS_TOKEN=/d' infra.env || true
        echo "MEMPHIS_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'memphis configured — shopos-events station created'
}
return this
