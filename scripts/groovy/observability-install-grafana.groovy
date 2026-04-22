def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install grafana observability/grafana/charts \
            --namespace grafana \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^GRAFANA_/d' infra.env || true"
    sh "sed -i '/^GRAFANA_URL=/d' infra.env 2>/dev/null || true; echo 'GRAFANA_URL=http://grafana-grafana.grafana.svc.cluster.local:3000' >> infra.env"
    echo 'grafana installed'
}
return this
