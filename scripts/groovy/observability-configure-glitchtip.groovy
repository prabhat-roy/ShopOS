def call() {
    sh """
        GT_URL=\$(grep '^GLITCHTIP_URL=' infra.env | cut -d= -f2)
        echo "Waiting for GlitchTip at \${GT_URL}..."
        until curl -sf "\${GT_URL}/api/0/" > /dev/null 2>&1; do sleep 10; done

        # Create superuser
        kubectl exec -n glitchtip deploy/glitchtip-glitchtip -- \
            python manage.py createsuperuser --noinput \
            --email admin@shopos.local --username admin || true

        # Create organisation and project via API
        TOKEN=\$(curl -sf -X POST "\${GT_URL}/api/0/auth/login/" \
            -H "Content-Type: application/json" \
            -d '{"login":"admin@shopos.local","password":"admin"}' \
            | grep -o '"token":"[^"]*"' | cut -d: -f2 | tr -d '"')

        curl -sf -X POST "\${GT_URL}/api/0/organizations/" \
            -H "Authorization: Bearer \${TOKEN}" \
            -H "Content-Type: application/json" \
            -d '{"name":"ShopOS","slug":"shopos"}' || true

        sed -i '/^GLITCHTIP_TOKEN=/d' infra.env || true
        echo "GLITCHTIP_TOKEN=\${TOKEN}" >> infra.env
    """
    echo 'glitchtip configured — shopos organisation created'
}
return this
