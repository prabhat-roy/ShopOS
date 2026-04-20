def call() {
    sh """
        helm upgrade --install kraken registry/charts/kraken \
            --namespace kraken \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://kraken-kraken.kraken.svc.cluster.local:16000'
    sh "sed -i '/^KRAKEN_/d' infra.env || true"
    sh "echo 'KRAKEN_URL=http://kraken-kraken.kraken.svc.cluster.local:16000' >> infra.env"

    echo 'kraken installed — ${url}'
}

return this
