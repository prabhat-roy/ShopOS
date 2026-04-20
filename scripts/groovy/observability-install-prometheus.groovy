def call() {
    sh """
        helm upgrade --install prometheus observability/prometheus/charts             --namespace prometheus             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^PROMETHEUS_/d' infra.env || true"
    sh "sed -i '/^PROMETHEUS_URL=/d' infra.env 2>/dev/null || true; echo 'PROMETHEUS_URL=http://prometheus-prometheus.prometheus.svc.cluster.local:9090' >> infra.env" 
    echo 'prometheus installed'
}
return this
