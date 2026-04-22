def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install zookeeper messaging/zookeeper/charts \
            --namespace zookeeper \
            --create-namespace \
            --set fullnameOverride=zookeeper \
            --set env.ZOOKEEPER_CLIENT_PORT=2181 \
            --set env.ZOOKEEPER_TICK_TIME=2000 \
            --set persistence.storageClass=${sc} \
            --wait --timeout 10m
    """
    sh "sed -i '/^ZOOKEEPER_/d' infra.env || true"
    sh "echo 'ZOOKEEPER_URL=zookeeper.zookeeper.svc.cluster.local:2181' >> infra.env"
    echo 'zookeeper installed'
}
return this
