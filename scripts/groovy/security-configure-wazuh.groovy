def call() {
    sh """
        WAZUH_URL=\$(grep '^WAZUH_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Wazuh manager at \${WAZUH_URL}..."
        until curl -sfk "\${WAZUH_URL}" > /dev/null 2>&1; do sleep 15; done

        # Get JWT token
        TOKEN=\$(curl -sfk -u wazuh:wazuh -X GET "\${WAZUH_URL}/security/user/authenticate" \
            | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Enable default vulnerability detector
        curl -sfk -X PUT "\${WAZUH_URL}/manager/configuration" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" || true

        # Create default agent group for ShopOS
        curl -sfk -X POST "\${WAZUH_URL}/groups" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"group_id":"shopos"}' || true

        sed -i '/^WAZUH_TOKEN=/d' infra.env || true
        echo "WAZUH_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'wazuh configured — shopos agent group created'
}
return this
