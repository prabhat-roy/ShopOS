def call() {
    sh """
        helm upgrade --install forgejo registry/charts/forgejo \
            --namespace forgejo \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://forgejo-forgejo.forgejo.svc.cluster.local:3000'
    sh "sed -i '/^FORGEJO_/d' infra.env || true"
    sh "sed -i '/^FORGEJO_URL=/d' infra.env 2>/dev/null || true; echo 'FORGEJO_URL=http://forgejo-forgejo.forgejo.svc.cluster.local:3000' >> infra.env" 
    sh "sed -i '/^FORGEJO_USER=/d' infra.env 2>/dev/null || true; echo 'FORGEJO_USER=forgejo_admin' >> infra.env" 
    sh "sed -i '/^FORGEJO_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'FORGEJO_PASSWORD=forgejo_admin' >> infra.env" 

    echo 'forgejo installed — ${url}'
}

return this
