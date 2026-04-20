def call() {
    sh """
        helm upgrade --install harbor registry/charts/harbor \
            --namespace harbor \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://harbor-harbor.harbor.svc.cluster.local:8080'
    sh "sed -i '/^HARBOR_/d' infra.env || true"
    sh "echo 'HARBOR_URL=http://harbor-harbor.harbor.svc.cluster.local:8080' >> infra.env"
    sh "echo 'HARBOR_USER=admin' >> infra.env"
    sh "echo 'HARBOR_PASSWORD=Harbor12345' >> infra.env"

    echo 'harbor installed — ${url}'
}

return this
