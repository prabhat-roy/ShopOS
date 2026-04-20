def call() {
    sh """
        INF_URL=\$(grep '^INFISICAL_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Infisical at \${INF_URL}..."
        until curl -sf "\${INF_URL}/api/status" > /dev/null 2>&1; do sleep 10; done

        # Create admin account
        curl -sf -X POST "\${INF_URL}/api/v1/auth/signup" \
            -H "Content-Type: application/json" \
            -d '{"email":"admin@shopos.local","firstName":"Admin","lastName":"ShopOS","password":"admin123"}' || true

        # Login and get token
        LOGIN=\$(curl -sf -X POST "\${INF_URL}/api/v1/auth/login1" \
            -H "Content-Type: application/json" \
            -d '{"email":"admin@shopos.local","clientPublicKey":""}')
        TOKEN=\$(echo "\${LOGIN}" | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create ShopOS organisation
        curl -sf -X POST "\${INF_URL}/api/v1/organization" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS"}' || true

        sed -i '/^INFISICAL_TOKEN=/d' infra.env || true
        echo "INFISICAL_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'infisical configured — admin account and ShopOS organisation created'
}
return this
