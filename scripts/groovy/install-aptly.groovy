def call() {
    sh """
        helm upgrade --install aptly registry/charts/aptly \
            --namespace aptly \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://aptly-aptly.aptly.svc.cluster.local:8080'
    sh "sed -i '/^APTLY_/d' infra.env || true"
    sh "echo 'APTLY_URL=http://aptly-aptly.aptly.svc.cluster.local:8080' >> infra.env"

    echo 'aptly installed — ${url}'
}

return this
