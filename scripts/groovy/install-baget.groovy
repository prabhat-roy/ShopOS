def call() {
    sh """
        helm upgrade --install baget registry/charts/baget \
            --namespace baget \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://baget-baget.baget.svc.cluster.local:8080'
    sh "sed -i '/^BAGET_/d' infra.env || true"
    sh "sed -i '/^BAGET_URL=/d' infra.env 2>/dev/null || true; echo 'BAGET_URL=http://baget-baget.baget.svc.cluster.local:8080' >> infra.env" 

    echo 'baget installed — ${url}'
}

return this
