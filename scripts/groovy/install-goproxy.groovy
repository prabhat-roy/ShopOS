def call() {
    sh """
        helm upgrade --install goproxy registry/charts/goproxy \
            --namespace goproxy \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://goproxy-goproxy.goproxy.svc.cluster.local:8081'
    sh "sed -i '/^GOPROXY_/d' infra.env || true"
    sh "sed -i '/^GOPROXY_URL=/d' infra.env 2>/dev/null || true; echo 'GOPROXY_URL=http://goproxy-goproxy.goproxy.svc.cluster.local:8081' >> infra.env" 

    echo 'goproxy installed — ${url}'
}

return this
