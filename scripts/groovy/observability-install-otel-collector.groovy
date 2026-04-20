def call() {
    sh """
        helm upgrade --install otel-collector observability/otel-collector/charts             --namespace otel-collector             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^OTEL_COLLECTOR_/d' infra.env || true"
    sh "sed -i '/^OTEL_COLLECTOR_URL=/d' infra.env 2>/dev/null || true; echo 'OTEL_COLLECTOR_URL=http://otel-collector-otel-collector.otel-collector.svc.cluster.local:4317' >> infra.env" 
    echo 'otel-collector installed'
}
return this
