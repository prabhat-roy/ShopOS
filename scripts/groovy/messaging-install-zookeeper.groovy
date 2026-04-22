def call() {
    sh """
        helm upgrade --install zookeeper messaging/zookeeper/charts \
            --namespace zookeeper \
            --create-namespace \
            --set env.ZOOKEEPER_CLIENT_PORT=2181 \
            --set env.ZOOKEEPER_TICK_TIME=2000 \
            --wait --timeout 10m
    """
    sh "sed -i '/^ZOOKEEPER_/d' infra.env || true"
    sh "sed -i '/^ZOOKEEPER_URL=/d' infra.env 2>/dev/null || true; echo 'ZOOKEEPER_URL=zookeeper-zookeeper.zookeeper.svc.cluster.local:2181' >> infra.env" 
    echo 'zookeeper installed'
}
return this
