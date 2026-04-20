def call() {
    sh """
        DD_URL=\$(grep '^DEFECTDOJO_URL=' infra.env | cut -d= -f2)
        echo "Waiting for DefectDojo at \${DD_URL}..."
        until curl -sf "\${DD_URL}/api/v2/user/" -u admin:admin > /dev/null 2>&1; do sleep 10; done

        # Get API token
        API_TOKEN=\$(curl -sf -X POST "\${DD_URL}/api/v2/api-token-auth/" \
            -H "Content-Type: application/json" \
            -d '{"username":"admin","password":"admin"}' \
            | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create ShopOS product type
        curl -sf -X POST "\${DD_URL}/api/v2/product_types/" \
            -H "Authorization: Token \${API_TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS","description":"ShopOS platform security findings"}' || true

        # Create ShopOS product
        PROD_TYPE_ID=\$(curl -sf "\${DD_URL}/api/v2/product_types/?name=ShopOS" \
            -H "Authorization: Token \${API_TOKEN}" \
            | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
        curl -sf -X POST "\${DD_URL}/api/v2/products/" \
            -H "Authorization: Token \${API_TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"ShopOS\",\"description\":\"ShopOS microservices\",\"prod_type\":\${PROD_TYPE_ID}}" || true

        # Create default engagement
        PROD_ID=\$(curl -sf "\${DD_URL}/api/v2/products/?name=ShopOS" \
            -H "Authorization: Token \${API_TOKEN}" \
            | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2)
        curl -sf -X POST "\${DD_URL}/api/v2/engagements/" \
            -H "Authorization: Token \${API_TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"CI Pipeline\",\"product\":\${PROD_ID},\"target_start\":\"\$(date +%Y-%m-%d)\",\"target_end\":\"\$(date -d '+1 year' +%Y-%m-%d)\",\"engagement_type\":\"CI/CD\",\"status\":\"In Progress\"}" || true

        sed -i '/^DEFECTDOJO_TOKEN=/d' infra.env || true
        echo "DEFECTDOJO_TOKEN=\${API_TOKEN}" >> infra.env
    """
    echo 'defectdojo configured — product, product type, and CI engagement created'
}
return this
