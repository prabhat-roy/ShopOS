def call() {
    sh """
        helm upgrade --install kellnr registry/charts/kellnr \
            --namespace kellnr \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://kellnr-kellnr.kellnr.svc.cluster.local:8080'
    sh "sed -i '/^KELLNR_/d' infra.env || true"
    sh "echo 'KELLNR_URL=http://kellnr-kellnr.kellnr.svc.cluster.local:8080' >> infra.env"
    sh "echo 'KELLNR_USER=admin' >> infra.env"
    sh "echo 'KELLNR_PASSWORD=admin' >> infra.env"

    echo 'kellnr installed — ${url}'
}

return this
