def call() {
    sh """
        helm upgrade --install distribution registry/charts/distribution \
            --namespace distribution \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://distribution-distribution.distribution.svc.cluster.local:5000'
    sh "sed -i '/^DISTRIBUTION_/d' infra.env || true"
    sh "sed -i '/^DISTRIBUTION_URL=/d' infra.env 2>/dev/null || true; echo 'DISTRIBUTION_URL=http://distribution-distribution.distribution.svc.cluster.local:5000' >> infra.env" 

    echo 'distribution installed — ${url}'
}

return this
