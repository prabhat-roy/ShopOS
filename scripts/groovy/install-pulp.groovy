def call() {
    sh """
        helm upgrade --install pulp registry/charts/pulp \
            --namespace pulp \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://pulp-pulp.pulp.svc.cluster.local:80'
    sh "sed -i '/^PULP_/d' infra.env || true"
    sh "echo 'PULP_URL=http://pulp-pulp.pulp.svc.cluster.local:80' >> infra.env"
    sh "echo 'PULP_USER=admin' >> infra.env"
    sh "echo 'PULP_PASSWORD=password' >> infra.env"

    echo 'pulp installed — ${url}'
}

return this
