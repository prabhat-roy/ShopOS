def call() {
    sh """
        helm upgrade --install quay registry/charts/quay \
            --namespace quay \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://quay-quay.quay.svc.cluster.local:8080'
    sh "sed -i '/^QUAY_/d' infra.env || true"
    sh "echo 'QUAY_URL=http://quay-quay.quay.svc.cluster.local:8080' >> infra.env"
    sh "echo 'QUAY_USER=quay' >> infra.env"
    sh "echo 'QUAY_PASSWORD=password' >> infra.env"

    echo 'quay installed — ${url}'
}

return this
