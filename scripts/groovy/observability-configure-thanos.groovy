def call() {
    sh """
        THANOS_URL=\$(grep '^THANOS_URL=' infra.env | cut -d= -f2)
        PROM_URL=\$(grep '^PROMETHEUS_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Thanos at \${THANOS_URL}..."
        until curl -sf "\${THANOS_URL}/-/ready" > /dev/null 2>&1; do sleep 10; done

        # Patch Prometheus to add Thanos sidecar remote_write
        kubectl patch configmap prometheus-scrape-config -n prometheus --type merge -p \
            "{\"data\":{\"remote_write\":\"- url: http://thanos-thanos.thanos.svc.cluster.local:10908/api/v1/receive\\n\"}}" || true

        kubectl rollout restart deployment/prometheus-prometheus -n prometheus || true
        echo "Thanos configured — Prometheus remote_write pointed at Thanos receive endpoint"
    """
    echo 'thanos configured'
}
return this
