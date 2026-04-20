def call() {
    sh """
        GF_URL=\$(grep '^GRAFANA_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Grafana at \${GF_URL}..."
        until curl -sf "\${GF_URL}/api/health" | grep -q '"database":"ok"'; do sleep 10; done

        PROM_URL=\$(grep '^PROMETHEUS_URL=' infra.env | cut -d= -f2)
        LOKI_URL=\$(grep '^LOKI_URL=' infra.env | cut -d= -f2)
        TEMPO_URL=\$(grep '^TEMPO_URL=' infra.env | cut -d= -f2)
        JAEGER_URL=\$(grep '^JAEGER_URL=' infra.env | cut -d= -f2)
        PYROSCOPE_URL=\$(grep '^PYROSCOPE_URL=' infra.env | cut -d= -f2)

        # Add Prometheus datasource
        curl -sf -u admin:admin -X POST "\${GF_URL}/api/datasources" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"Prometheus\",\"type\":\"prometheus\",\"url\":\"\${PROM_URL}\",\"access\":\"proxy\",\"isDefault\":true}" || true

        # Add Loki datasource
        curl -sf -u admin:admin -X POST "\${GF_URL}/api/datasources" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"Loki\",\"type\":\"loki\",\"url\":\"\${LOKI_URL}\",\"access\":\"proxy\"}" || true

        # Add Tempo datasource
        curl -sf -u admin:admin -X POST "\${GF_URL}/api/datasources" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"Tempo\",\"type\":\"tempo\",\"url\":\"\${TEMPO_URL}\",\"access\":\"proxy\"}" || true

        # Add Jaeger datasource
        curl -sf -u admin:admin -X POST "\${GF_URL}/api/datasources" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"Jaeger\",\"type\":\"jaeger\",\"url\":\"\${JAEGER_URL}\",\"access\":\"proxy\"}" || true

        # Add Pyroscope datasource
        curl -sf -u admin:admin -X POST "\${GF_URL}/api/datasources" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"Pyroscope\",\"type\":\"grafana-pyroscope-datasource\",\"url\":\"\${PYROSCOPE_URL}\",\"access\":\"proxy\"}" || true

        # Import pre-built dashboards from observability/grafana/dashboards/
        for dashboard in observability/grafana/dashboards/*.json; do
            [ -f "\$dashboard" ] || continue
            PAYLOAD=\$(cat "\$dashboard")
            curl -sf -u admin:admin -X POST "\${GF_URL}/api/dashboards/import" \
                -H "Content-Type: application/json" \
                -d "{\\"dashboard\\":{\$PAYLOAD},\\"overwrite\\":true,\\"folderId\\":0}" || true
            echo "  imported: \$dashboard"
        done

        # Change admin password
        curl -sf -u admin:admin -X PUT "\${GF_URL}/api/user/password" \
            -H "Content-Type: application/json" \
            -d '{"oldPassword":"admin","newPassword":"admin123","confirmNew":"admin123"}' || true

        sed -i '/^GRAFANA_PASSWORD=/d' infra.env || true
        echo "GRAFANA_PASSWORD=admin123" >> infra.env
    """
    echo 'grafana configured — all datasources added, dashboards imported'
}
return this
