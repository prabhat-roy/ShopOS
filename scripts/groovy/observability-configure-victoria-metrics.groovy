def call() {
    sh '''
        echo "=== Configure VictoriaMetrics ==="

        kubectl rollout status deploy/victoria-metrics -n victoria-metrics --timeout=120s || true

        VM_IP=$(kubectl get svc victoria-metrics -n victoria-metrics \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "victoria-metrics.victoria-metrics.svc.cluster.local")
        VM_URL="http://${VM_IP}:8428"

        sed -i '/^VICTORIA_METRICS_URL=/d' infra.env
        echo "VICTORIA_METRICS_URL=${VM_URL}" >> infra.env

        # Register as datasource in Grafana (Prometheus-compatible)
        GRAFANA_URL=$(grep '^GRAFANA_URL=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        GRAFANA_PASS=$(grep '^GRAFANA_ADMIN_PASS=' infra.env 2>/dev/null | cut -d= -f2 || echo "admin")
        if [ -n "$GRAFANA_URL" ]; then
            curl -sf -X POST "${GRAFANA_URL}/api/datasources" \
                -u "admin:${GRAFANA_PASS}" \
                -H "Content-Type: application/json" \
                -d "{\"name\":\"VictoriaMetrics\",\"type\":\"prometheus\",\"url\":\"${VM_URL}\",\"access\":\"proxy\",\"isDefault\":false}" \
                2>/dev/null || true
            echo "  VictoriaMetrics datasource registered in Grafana"
        fi

        echo "VictoriaMetrics configured. URL written to infra.env."
    '''
}
return this
