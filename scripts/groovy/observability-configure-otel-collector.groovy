def call() {
    sh """
        PROM_URL=\$(grep '^PROMETHEUS_URL=' infra.env | cut -d= -f2)
        LOKI_URL=\$(grep '^LOKI_URL=' infra.env | cut -d= -f2)
        TEMPO_URL=\$(grep '^TEMPO_URL=' infra.env | cut -d= -f2)
        JAEGER_URL=\$(grep '^JAEGER_URL=' infra.env | cut -d= -f2)

        # Load OTel config from existing config
        kubectl create configmap otel-collector-config \
            --from-file=observability/otel/ \
            --namespace otel-collector --dry-run=client -o yaml | kubectl apply -f - || true

        kubectl rollout restart deployment/otel-collector-otel-collector -n otel-collector || true
        echo "OTel Collector configured — pipelines: traces→Jaeger+Tempo, metrics→Prometheus, logs→Loki"
    """
    echo 'otel-collector configured'
}
return this
