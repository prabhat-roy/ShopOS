def call() {
    sh """
        KC_URL=\$(grep '^KEYCLOAK_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Keycloak at \${KC_URL}..."
        until curl -sf "\${KC_URL}/health/ready" > /dev/null 2>&1; do sleep 10; done

        # Get admin token
        TOKEN=\$(curl -sf -X POST "\${KC_URL}/realms/master/protocol/openid-connect/token" \
            -d "client_id=admin-cli&username=admin&password=admin&grant_type=password" \
            | grep -o '"access_token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create shopos realm
        curl -sf -X POST "\${KC_URL}/admin/realms" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"realm":"shopos","enabled":true,"displayName":"ShopOS","registrationAllowed":false}' || true

        # Create shopos-app client
        curl -sf -X POST "\${KC_URL}/admin/realms/shopos/clients" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"clientId":"shopos-app","enabled":true,"publicClient":false,"standardFlowEnabled":true,"serviceAccountsEnabled":true}' || true

        sed -i '/^KEYCLOAK_REALM=/d' infra.env || true
        echo "KEYCLOAK_REALM=shopos" >> infra.env
    """
    echo 'keycloak configured — shopos realm and client created'
}
return this
