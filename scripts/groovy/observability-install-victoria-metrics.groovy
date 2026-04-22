def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install victoria-metrics observability/victoria-metrics/charts \
            --namespace victoria-metrics \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^VICTORIA_METRICS_/d' infra.env || true"
    sh "sed -i '/^VICTORIA_METRICS_URL=/d' infra.env 2>/dev/null || true; echo 'VICTORIA_METRICS_URL=http://victoria-metrics-victoria-metrics.victoria-metrics.svc.cluster.local:8428' >> infra.env"
    echo 'victoria-metrics installed'
}
return this
