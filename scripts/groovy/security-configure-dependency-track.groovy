def call() {
    sh """
        DT_URL=\$(grep '^DEPENDENCY_TRACK_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Dependency Track at \${DT_URL}..."
        until curl -sf "\${DT_URL}/api/version" > /dev/null 2>&1; do sleep 10; done

        # Get API key for default admin team
        API_KEY=\$(curl -sf -u admin:admin "\${DT_URL}/api/v1/team" \
            | grep -o '"apiKeys":\\[{"key":"[^"]*"' | grep -o '"key":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create ShopOS project
        curl -sf -X PUT "\${DT_URL}/api/v1/project" \
            -H "X-Api-Key: \${API_KEY}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS","version":"1.0.0","description":"ShopOS platform SBOM tracking","active":true}' || true

        sed -i '/^DEPENDENCY_TRACK_API_KEY=/d' infra.env || true
        echo "DEPENDENCY_TRACK_API_KEY=\${API_KEY}" >> infra.env
    """
    echo 'dependency-track configured — ShopOS project created, API key written to infra.env'
}
return this
