def call() {
    sh """
        helm upgrade --install zot registry/charts/zot \
            --namespace zot \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://zot-zot.zot.svc.cluster.local:5080'
    sh "sed -i '/^ZOT_/d' infra.env || true"
    sh "echo 'ZOT_URL=http://zot-zot.zot.svc.cluster.local:5080' >> infra.env"

    echo 'zot installed — ${url}'
}

return this
