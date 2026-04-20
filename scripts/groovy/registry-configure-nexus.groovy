def call() {
    sh '''
        echo "=== Configure Nexus ==="

        NEXUS_IP=$(kubectl get svc nexus -n nexus \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "nexus.nexus.svc.cluster.local")
        NEXUS_URL="http://${NEXUS_IP}:8081"

        # Wait for Nexus
        for i in $(seq 1 24); do
            curl -sf "${NEXUS_URL}/service/rest/v1/status" >/dev/null 2>&1 && break
            echo "  Waiting for Nexus... ($i/24)"
            sleep 10
        done

        # Retrieve initial admin password
        NEXUS_PASS=$(kubectl exec -n nexus deploy/nexus -- \
            cat /nexus-data/admin.password 2>/dev/null || echo "admin123")

        curl_nexus() {
            curl -sf -u "admin:${NEXUS_PASS}" "$@"
        }

        # Helper — create hosted repo if not exists
        create_hosted() {
            local name=$1 format=$2
            curl_nexus -X POST "${NEXUS_URL}/service/rest/v1/repositories/${format}/hosted" \
                -H "Content-Type: application/json" \
                -d "{\"name\":\"${name}\",\"online\":true,\"storage\":{\"blobStoreName\":\"default\",\"strictContentTypeValidation\":true,\"writePolicy\":\"allow_once\"}}" \
                2>/dev/null || true
        }

        # Create hosted repos for each language ecosystem
        create_hosted "shopos-maven"  "maven2"
        create_hosted "shopos-npm"    "npm"
        create_hosted "shopos-pypi"   "pypi"
        create_hosted "shopos-go"     "go"
        create_hosted "shopos-docker" "docker"
        create_hosted "shopos-helm"   "helm"
        create_hosted "shopos-cargo"  "cargo"
        create_hosted "shopos-nuget"  "nuget"

        sed -i '/^NEXUS_URL=/d' infra.env
        echo "NEXUS_URL=${NEXUS_URL}" >> infra.env

        echo "Nexus hosted repositories configured for all language ecosystems."
    '''
}
return this
