def call() {
    sh """
        helm upgrade --install quetz registry/charts/quetz \
            --namespace quetz \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://quetz-quetz.quetz.svc.cluster.local:8000'
    sh "sed -i '/^QUETZ_/d' infra.env || true"
    sh "echo 'QUETZ_URL=http://quetz-quetz.quetz.svc.cluster.local:8000' >> infra.env"

    echo 'quetz installed — ${url}'
}

return this
