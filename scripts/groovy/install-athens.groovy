def call() {
    sh """
        helm upgrade --install athens registry/charts/athens \
            --namespace athens \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://athens-athens.athens.svc.cluster.local:3000'
    sh "sed -i '/^ATHENS_/d' infra.env || true"
    sh "sed -i '/^ATHENS_URL=/d' infra.env 2>/dev/null || true; echo 'ATHENS_URL=http://athens-athens.athens.svc.cluster.local:3000' >> infra.env" 

    echo 'athens installed — ${url}'
}

return this
