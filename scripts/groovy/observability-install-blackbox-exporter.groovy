def call() {
    sh """
        helm upgrade --install blackbox-exporter observability/blackbox-exporter/charts             --namespace blackbox-exporter             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^BLACKBOX_/d' infra.env || true"
    sh "echo 'BLACKBOX_URL=http://blackbox-exporter-blackbox-exporter.blackbox-exporter.svc.cluster.local:9115' >> infra.env"
    echo 'blackbox-exporter installed'
}
return this
