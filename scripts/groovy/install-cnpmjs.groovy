def call() {
    sh """
        helm upgrade --install cnpmjs registry/charts/cnpmjs \
            --namespace cnpmjs \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://cnpmjs-cnpmjs.cnpmjs.svc.cluster.local:7001'
    sh "sed -i '/^CNPMJS_/d' infra.env || true"
    sh "sed -i '/^CNPMJS_URL=/d' infra.env 2>/dev/null || true; echo 'CNPMJS_URL=http://cnpmjs-cnpmjs.cnpmjs.svc.cluster.local:7001' >> infra.env" 

    echo 'cnpmjs installed — ${url}'
}

return this
