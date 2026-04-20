def call() {
    sh """
        helm upgrade --install grafana observability/grafana/charts             --namespace grafana             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^GRAFANA_/d' infra.env || true"
    sh "echo 'GRAFANA_URL=http://grafana-grafana.grafana.svc.cluster.local:3000' >> infra.env"
    echo 'grafana installed'
}
return this
