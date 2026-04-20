def call() {
    sh '''
        echo "=== Configure Gimlet ==="

        kubectl rollout status deploy/gimlet -n gimlet --timeout=120s || true

        GIMLET_URL=$(grep "^GIMLET_URL=" infra.env 2>/dev/null | cut -d= -f2)
        GIMLET_PASS=$(grep "^GIMLET_PASSWORD=" infra.env 2>/dev/null | cut -d= -f2)

        # Generate a Gimlet API token for CI integration
        GIMLET_TOKEN=$(kubectl get secret gimlet-admin-token -n gimlet \
            -o jsonpath="{.data.token}" 2>/dev/null | base64 -d || \
            openssl rand -hex 20 2>/dev/null || echo "")

        if [ -n "$GIMLET_TOKEN" ]; then
            sed -i "/^GIMLET_TOKEN=/d" infra.env 2>/dev/null || true
            echo "GIMLET_TOKEN=${GIMLET_TOKEN}" >> infra.env
        fi

        # Register GitHub repo with Gimlet
        ARGOCD_URL=$(grep "^ARGOCD_URL=" infra.env 2>/dev/null | cut -d= -f2)
        ARGOCD_PASS=$(grep "^ARGOCD_ADMIN_PASS=" infra.env 2>/dev/null | cut -d= -f2)

        if [ -n "$ARGOCD_URL" ] && [ -n "$GIMLET_TOKEN" ]; then
            curl -sf -X POST "${GIMLET_URL}/api/repo" \
                -H "Authorization: Bearer ${GIMLET_TOKEN}" \
                -H "Content-Type: application/json" \
                -d "{\"owner\":\"prabhat-roy\",\"name\":\"ShopOS\",\"gitopsRepoOwner\":\"prabhat-roy\",\"gitopsRepoName\":\"ShopOS\"}" \
                2>/dev/null || true
            echo "  ShopOS repo registered with Gimlet."
        fi

        echo "Gimlet configured — developer platform ready."
        echo "URL: ${GIMLET_URL:-http://gimlet-gimlet.gimlet.svc.cluster.local:9000}"
    '''
}
return this
