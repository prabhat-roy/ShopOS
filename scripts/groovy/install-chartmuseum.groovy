def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install chartmuseum registry/charts/chartmuseum \
            --namespace chartmuseum \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """

    def url = 'http://chartmuseum-chartmuseum.chartmuseum.svc.cluster.local:8080'
    sh "sed -i '/^CHARTMUSEUM_/d' infra.env || true"
    sh "sed -i '/^CHARTMUSEUM_URL=/d' infra.env 2>/dev/null || true; echo 'CHARTMUSEUM_URL=http://chartmuseum-chartmuseum.chartmuseum.svc.cluster.local:8080' >> infra.env" 

    echo 'chartmuseum installed — ${url}'
}

return this
