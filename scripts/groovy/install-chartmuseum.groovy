def call() {
    sh """
        helm upgrade --install chartmuseum registry/charts/chartmuseum \
            --namespace chartmuseum \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://chartmuseum-chartmuseum.chartmuseum.svc.cluster.local:8080'
    sh "sed -i '/^CHARTMUSEUM_/d' infra.env || true"
    sh "echo 'CHARTMUSEUM_URL=http://chartmuseum-chartmuseum.chartmuseum.svc.cluster.local:8080' >> infra.env"

    echo 'chartmuseum installed — ${url}'
}

return this
