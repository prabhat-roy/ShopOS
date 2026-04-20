def call() {
    sh """
        UK_URL=\$(grep '^UPTIME_KUMA_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Uptime Kuma at \${UK_URL}..."
        until curl -sf "\${UK_URL}" > /dev/null 2>&1; do sleep 10; done

        # Setup admin account via setup API
        curl -sf -X POST "\${UK_URL}/api/setup" \
            -H "Content-Type: application/json" \
            -d '{"username":"admin","password":"admin123"}' || true

        # Login to get token
        TOKEN=\$(curl -sf -X POST "\${UK_URL}/api/login" \
            -H "Content-Type: application/json" \
            -d '{"username":"admin","password":"admin123"}' \
            | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        sed -i '/^UPTIME_KUMA_TOKEN=/d' infra.env || true
        echo "UPTIME_KUMA_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'uptime-kuma configured — admin account created'
}
return this
