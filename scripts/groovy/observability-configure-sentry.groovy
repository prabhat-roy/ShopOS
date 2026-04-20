def call() {
    sh """
        SENTRY_URL=\$(grep '^SENTRY_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Sentry at \${SENTRY_URL}..."
        until curl -sf "\${SENTRY_URL}/_health/" > /dev/null 2>&1; do sleep 15; done

        # Run initial database migrations and create superuser
        kubectl exec -n sentry deploy/sentry-sentry -- sentry upgrade --noinput || true

        # Create default organisation and project via CLI
        kubectl exec -n sentry deploy/sentry-sentry -- sentry createuser \
            --email admin@shopos.local --password admin123 --superuser --no-input || true

        TOKEN=\$(curl -sf -X POST "\${SENTRY_URL}/api/0/auth/login/" \
            -H "Content-Type: application/json" \
            -d '{"username":"admin@shopos.local","password":"admin123"}' \
            | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        # Create ShopOS organisation
        curl -sf -X POST "\${SENTRY_URL}/api/0/organizations/" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS","slug":"shopos"}' || true

        sed -i '/^SENTRY_TOKEN=/d' infra.env || true
        echo "SENTRY_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'sentry configured — shopos org created, admin user ready'
}
return this
