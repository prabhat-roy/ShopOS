def call() {
    sh """
        helm upgrade --install baget registry/charts/baget \
            --namespace baget \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://baget-baget.baget.svc.cluster.local:8080'
    sh "sed -i '/^BAGET_/d' infra.env || true"
    sh "echo 'BAGET_URL=http://baget-baget.baget.svc.cluster.local:8080' >> infra.env"

    echo 'baget installed — ${url}'
}

return this
