def call() {
    sh '''
        echo "=== Configure OpenSearch Dashboards ==="

        kubectl rollout status deploy/opensearch-dashboards -n opensearch-dashboards --timeout=180s || true

        OSD_IP=$(kubectl get svc opensearch-dashboards -n opensearch-dashboards \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "opensearch-dashboards.opensearch-dashboards.svc.cluster.local")
        OSD_URL="http://${OSD_IP}:5601"

        sed -i '/^OPENSEARCH_DASHBOARDS_URL=/d' infra.env
        echo "OPENSEARCH_DASHBOARDS_URL=${OSD_URL}" >> infra.env

        # Wait for dashboards readiness
        for i in $(seq 1 24); do
            STATUS=$(curl -sf -o /dev/null -w "%{http_code}" "${OSD_URL}/api/status" 2>/dev/null || echo "000")
            [ "$STATUS" = "200" ] && break
            echo "  Waiting for OpenSearch Dashboards... ($i/24)"
            sleep 10
        done

        # Create index pattern for ShopOS logs
        curl -sf -X POST "${OSD_URL}/api/saved_objects/index-pattern/shopos-logs-*" \
            -H "osd-xsrf: true" \
            -H "Content-Type: application/json" \
            -d '{"attributes":{"title":"shopos-logs-*","timeFieldName":"@timestamp"}}' \
            2>/dev/null || true

        echo "OpenSearch Dashboards index pattern created. URL written to infra.env."
    '''
}
return this
