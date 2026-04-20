def call() {
    sh """
        helm upgrade --install devpi registry/charts/devpi \
            --namespace devpi \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://devpi-devpi.devpi.svc.cluster.local:3141'
    sh "sed -i '/^DEVPI_/d' infra.env || true"
    sh "echo 'DEVPI_URL=http://devpi-devpi.devpi.svc.cluster.local:3141' >> infra.env"
    sh "echo 'DEVPI_USER=root' >> infra.env"

    echo 'devpi installed — ${url}'
}

return this
