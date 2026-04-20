def call() {
    sh '''
        echo "=== Configure Grafana Pyroscope ==="

        kubectl rollout status deploy/pyroscope -n pyroscope --timeout=120s || true

        PYROSCOPE_IP=$(kubectl get svc pyroscope -n pyroscope \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "pyroscope.pyroscope.svc.cluster.local")
        PYROSCOPE_URL="http://${PYROSCOPE_IP}:4040"

        sed -i '/^PYROSCOPE_URL=/d' infra.env
        echo "PYROSCOPE_URL=${PYROSCOPE_URL}" >> infra.env

        # Register Pyroscope as a datasource in Grafana
        GRAFANA_URL=$(grep '^GRAFANA_URL=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        GRAFANA_PASS=$(grep '^GRAFANA_ADMIN_PASS=' infra.env 2>/dev/null | cut -d= -f2 || echo "admin")
        if [ -n "$GRAFANA_URL" ]; then
            curl -sf -X POST "${GRAFANA_URL}/api/datasources" \
                -u "admin:${GRAFANA_PASS}" \
                -H "Content-Type: application/json" \
                -d "{\"name\":\"Pyroscope\",\"type\":\"grafana-pyroscope-datasource\",\"url\":\"${PYROSCOPE_URL}\",\"access\":\"proxy\",\"isDefault\":false}" \
                2>/dev/null || true
            echo "  Pyroscope datasource registered in Grafana"
        fi

        echo "Pyroscope configured. URL written to infra.env."
    '''
}
return this
