def call() {
    sh """
        helm upgrade --install alexandrie registry/charts/alexandrie \
            --namespace alexandrie \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://alexandrie-alexandrie.alexandrie.svc.cluster.local:3000'
    sh "sed -i '/^ALEXANDRIE_/d' infra.env || true"
    sh "sed -i '/^ALEXANDRIE_URL=/d' infra.env 2>/dev/null || true; echo 'ALEXANDRIE_URL=http://alexandrie-alexandrie.alexandrie.svc.cluster.local:3000' >> infra.env" 

    echo 'alexandrie installed — ${url}'
}

return this
