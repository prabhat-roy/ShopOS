def call() {
    sh """
        helm upgrade --install dragonfly registry/charts/dragonfly \
            --namespace dragonfly \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://dragonfly-dragonfly.dragonfly.svc.cluster.local:8080'
    sh "sed -i '/^DRAGONFLY_/d' infra.env || true"
    sh "sed -i '/^DRAGONFLY_URL=/d' infra.env 2>/dev/null || true; echo 'DRAGONFLY_URL=http://dragonfly-dragonfly.dragonfly.svc.cluster.local:8080' >> infra.env" 

    echo 'dragonfly installed — ${url}'
}

return this
