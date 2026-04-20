def call() {
    sh '''
        echo "=== Configure Grafana Tempo ==="

        kubectl rollout status deploy/tempo -n tempo --timeout=120s || true

        TEMPO_IP=$(kubectl get svc tempo -n tempo \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "tempo.tempo.svc.cluster.local")

        sed -i '/^TEMPO_URL=/d' infra.env
        echo "TEMPO_URL=http://${TEMPO_IP}:3100" >> infra.env

        # Configure Tempo as a datasource in Grafana if available
        GRAFANA_URL=$(grep '^GRAFANA_URL=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        GRAFANA_PASS=$(grep '^GRAFANA_ADMIN_PASS=' infra.env 2>/dev/null | cut -d= -f2 || echo "admin")
        if [ -n "$GRAFANA_URL" ]; then
            curl -sf -X POST "${GRAFANA_URL}/api/datasources" \
                -u "admin:${GRAFANA_PASS}" \
                -H "Content-Type: application/json" \
                -d "{\"name\":\"Tempo\",\"type\":\"tempo\",\"url\":\"http://${TEMPO_IP}:3100\",\"access\":\"proxy\",\"isDefault\":false}" \
                2>/dev/null || true
            echo "  Tempo datasource registered in Grafana"
        fi

        echo "Tempo configured — URL written to infra.env."
    '''
}
return this
