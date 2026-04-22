def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install zot registry/charts/zot \
            --namespace zot \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """

    def url = 'http://zot-zot.zot.svc.cluster.local:5080'
    sh "sed -i '/^ZOT_/d' infra.env || true"
    sh "sed -i '/^ZOT_URL=/d' infra.env 2>/dev/null || true; echo 'ZOT_URL=http://zot-zot.zot.svc.cluster.local:5080' >> infra.env" 

    echo 'zot installed — ${url}'
}

return this
