def call() {
    sh """
        helm upgrade --install gogs registry/charts/gogs \
            --namespace gogs \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://gogs-gogs.gogs.svc.cluster.local:3000'
    sh "sed -i '/^GOGS_/d' infra.env || true"
    sh "echo 'GOGS_URL=http://gogs-gogs.gogs.svc.cluster.local:3000' >> infra.env"

    echo 'gogs installed — ${url}'
}

return this
