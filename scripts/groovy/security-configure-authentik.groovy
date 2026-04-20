def call() {
    sh """
        AK_URL=\$(grep '^AUTHENTIK_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Authentik at \${AK_URL}..."
        until curl -sf "\${AK_URL}/-/health/ready/" > /dev/null 2>&1; do sleep 10; done

        # Get admin token via bootstrap token
        TOKEN=\$(curl -sf -X POST "\${AK_URL}/api/v3/core/tokens/" \
            -H "Content-Type: application/json" \
            -d '{"identifier":"jenkins-token","intent":"api","user":1}' \
            | grep -o '"key":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create shopos application
        curl -sf -X POST "\${AK_URL}/api/v3/core/applications/" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS","slug":"shopos","open_in_new_tab":false}' || true

        sed -i '/^AUTHENTIK_TOKEN=/d' infra.env || true
        echo "AUTHENTIK_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'authentik configured — shopos application created'
}
return this
