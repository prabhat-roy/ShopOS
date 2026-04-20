def call() {
    sh '''
        echo "=== Configure Harbor ==="

        HARBOR_URL=$(grep '^HARBOR_URL=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        HARBOR_USER=$(grep '^HARBOR_USER=' infra.env 2>/dev/null | cut -d= -f2 || echo "admin")
        HARBOR_PASS=$(grep '^HARBOR_PASSWORD=' infra.env 2>/dev/null | cut -d= -f2 || echo "")

        if [ -z "$HARBOR_URL" ]; then
            HARBOR_IP=$(kubectl get svc harbor -n harbor \
                -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "harbor.harbor.svc.cluster.local")
            HARBOR_URL="http://${HARBOR_IP}"
            sed -i '/^HARBOR_URL=/d' infra.env
            echo "HARBOR_URL=${HARBOR_URL}" >> infra.env
        fi

        # Wait for Harbor to be ready
        for i in $(seq 1 18); do
            curl -sf "${HARBOR_URL}/api/v2.0/ping" >/dev/null 2>&1 && break
            echo "  Waiting for Harbor... ($i/18)"
            sleep 10
        done

        AUTH="-u ${HARBOR_USER}:${HARBOR_PASS}"

        # Create ShopOS project for each domain
        for project in commerce platform identity catalog supply-chain financial \\
                       customer-experience communications content analytics-ai b2b \\
                       integrations affiliate shopos; do
            curl -sf -X POST "${HARBOR_URL}/api/v2.0/projects" $AUTH \
                -H "Content-Type: application/json" \
                -d "{\"project_name\":\"${project}\",\"public\":false,\"metadata\":{\"auto_scan\":\"true\",\"severity\":\"high\"}}" \
                2>/dev/null || true
        done

        # Enable Trivy scanner as default
        SCANNER_ID=$(curl -sf "${HARBOR_URL}/api/v2.0/scanners" $AUTH 2>/dev/null \
            | python3 -c "import json,sys; d=json.load(sys.stdin); print(next((s['uuid'] for s in d if 'trivy' in s.get('name','').lower()), ''))" 2>/dev/null || echo "")
        if [ -n "$SCANNER_ID" ]; then
            curl -sf -X PATCH "${HARBOR_URL}/api/v2.0/scanners/${SCANNER_ID}" $AUTH \
                -H "Content-Type: application/json" \
                -d '{"is_default":true}' 2>/dev/null || true
            echo "  Trivy set as default scanner"
        fi

        # Create robot account for CI
        ROBOT_RESP=$(curl -sf -X POST "${HARBOR_URL}/api/v2.0/robots" $AUTH \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"ci-robot\",\"duration\":-1,\"level\":\"system\",\"permissions\":[{\"kind\":\"project\",\"namespace\":\"*\",\"access\":[{\"resource\":\"repository\",\"action\":\"push\"},{\"resource\":\"repository\",\"action\":\"pull\"},{\"resource\":\"artifact\",\"action\":\"delete\"}]}]}" \
            2>/dev/null || echo "")
        ROBOT_SECRET=$(echo "$ROBOT_RESP" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('secret',''))" 2>/dev/null || echo "")
        if [ -n "$ROBOT_SECRET" ]; then
            sed -i '/^HARBOR_ROBOT_SECRET=/d' infra.env
            echo "HARBOR_ROBOT_SECRET=${ROBOT_SECRET}" >> infra.env
            echo "  CI robot account created — secret written to infra.env"
        fi

        echo "Harbor projects, scanner, and CI robot account configured."
    '''
}
return this
