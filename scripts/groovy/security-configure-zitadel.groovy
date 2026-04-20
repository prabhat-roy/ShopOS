def call() {
    sh """
        ZI_URL=\$(grep '^ZITADEL_URL=' infra.env | cut -d= -f2)
        echo "Waiting for ZITADEL at \${ZI_URL}..."
        until curl -sf "\${ZI_URL}/debug/healthz" > /dev/null 2>&1; do sleep 10; done

        # Get service account token
        TOKEN=\$(curl -sf -X POST "\${ZI_URL}/oauth/v2/token" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "grant_type=client_credentials&client_id=zitadel-admin-sa&client_secret=&scope=openid+urn:zitadel:iam:org:project:id:zitadel:aud" \
            | grep -o '"access_token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create ShopOS project
        curl -sf -X POST "\${ZI_URL}/management/v1/projects" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS","projectRoleAssertion":true,"projectRoleCheck":true}' || true

        sed -i '/^ZITADEL_TOKEN=/d' infra.env || true
        echo "ZITADEL_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'zitadel configured — ShopOS project created'
}
return this
