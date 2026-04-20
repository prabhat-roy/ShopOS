def call() {
    sh """
        helm upgrade --install pulp registry/charts/pulp \
            --namespace pulp \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://pulp-pulp.pulp.svc.cluster.local:80'
    sh "sed -i '/^PULP_/d' infra.env || true"
    sh "sed -i '/^PULP_URL=/d' infra.env 2>/dev/null || true; echo 'PULP_URL=http://pulp-pulp.pulp.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^PULP_USER=/d' infra.env 2>/dev/null || true; echo 'PULP_USER=admin' >> infra.env" 
    sh "sed -i '/^PULP_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'PULP_PASSWORD=password' >> infra.env" 

    echo 'pulp installed — ${url}'
}

return this
