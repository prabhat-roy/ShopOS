def call() {
    sh '''
        echo "=== Configure Kibana ==="

        kubectl rollout status deploy/kibana -n kibana --timeout=180s || true

        KIBANA_IP=$(kubectl get svc kibana -n kibana \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "kibana.kibana.svc.cluster.local")
        KIBANA_URL="http://${KIBANA_IP}:5601"

        sed -i '/^KIBANA_URL=/d' infra.env
        echo "KIBANA_URL=${KIBANA_URL}" >> infra.env

        # Wait for Kibana to be ready
        for i in $(seq 1 24); do
            STATUS=$(curl -sf -o /dev/null -w "%{http_code}" "${KIBANA_URL}/api/status" 2>/dev/null || echo "000")
            [ "$STATUS" = "200" ] && break
            echo "  Waiting for Kibana... ($i/24) HTTP ${STATUS}"
            sleep 10
        done

        # Create index pattern for ShopOS logs
        curl -sf -X POST "${KIBANA_URL}/api/saved_objects/index-pattern/shopos-logs-*" \
            -H "kbn-xsrf: true" \
            -H "Content-Type: application/json" \
            -d '{"attributes":{"title":"shopos-logs-*","timeFieldName":"@timestamp"}}' \
            2>/dev/null || true

        curl -sf -X POST "${KIBANA_URL}/api/saved_objects/index-pattern/shopos-k8s-logs-*" \
            -H "kbn-xsrf: true" \
            -H "Content-Type: application/json" \
            -d '{"attributes":{"title":"shopos-k8s-logs-*","timeFieldName":"@timestamp"}}' \
            2>/dev/null || true

        echo "Kibana index patterns created. URL written to infra.env."
    '''
}
return this
