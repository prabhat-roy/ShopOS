def call() {
    sh '''
        echo "=== Configure Gitea ==="

        GITEA_IP=$(kubectl get svc gitea-http -n gitea \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "gitea-http.gitea.svc.cluster.local")
        GITEA_URL="http://${GITEA_IP}:3000"
        GITEA_ADMIN_USER=$(grep '^GITEA_ADMIN_USER=' infra.env 2>/dev/null | cut -d= -f2 || echo "gitea")
        GITEA_ADMIN_PASS=$(grep '^GITEA_ADMIN_PASS=' infra.env 2>/dev/null | cut -d= -f2 || echo "gitea123")

        # Wait for Gitea
        for i in $(seq 1 18); do
            curl -sf "${GITEA_URL}/api/v1/version" >/dev/null 2>&1 && break
            echo "  Waiting for Gitea... ($i/18)"
            sleep 10
        done

        gitea_api() {
            curl -sf -u "${GITEA_ADMIN_USER}:${GITEA_ADMIN_PASS}" \
                -H "Content-Type: application/json" "$@"
        }

        # Create shopos organisation
        gitea_api -X POST "${GITEA_URL}/api/v1/orgs" \
            -d '{"username":"shopos","full_name":"ShopOS","description":"ShopOS Enterprise Platform","visibility":"private"}' \
            2>/dev/null || true

        # Create main platform repo
        gitea_api -X POST "${GITEA_URL}/api/v1/orgs/shopos/repos" \
            -d '{"name":"enterprise-platform","description":"ShopOS enterprise-grade commerce platform","private":true,"default_branch":"main","auto_init":true}' \
            2>/dev/null || true

        # Generate API token for CI
        TOKEN_RESP=$(gitea_api -X POST "${GITEA_URL}/api/v1/users/${GITEA_ADMIN_USER}/tokens" \
            -d '{"name":"ci-token","scopes":["write:repository","read:user"]}' 2>/dev/null || echo "")
        GITEA_TOKEN=$(echo "$TOKEN_RESP" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('sha1',''))" 2>/dev/null || echo "")
        if [ -n "$GITEA_TOKEN" ]; then
            sed -i '/^GITEA_TOKEN=/d; /^GITEA_URL=/d' infra.env
            echo "GITEA_URL=${GITEA_URL}" >> infra.env
            echo "GITEA_TOKEN=${GITEA_TOKEN}" >> infra.env
            echo "  Gitea CI token written to infra.env"
        fi

        echo "Gitea organisation, repo, and CI token configured."
    '''
}
return this
