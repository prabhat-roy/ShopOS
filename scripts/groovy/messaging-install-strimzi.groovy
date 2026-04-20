def call() {
    sh """
        helm upgrade --install strimzi messaging/strimzi/charts \
            --namespace strimzi \
            --create-namespace \
            --set env.STRIMZI_NAMESPACE=kafka \
            --wait --timeout 5m
    """
    sh "sed -i '/^STRIMZI_/d' infra.env || true"
    sh "echo 'STRIMZI_URL=http://strimzi-strimzi.strimzi.svc.cluster.local:8080' >> infra.env"
    echo 'strimzi installed'
}
return this
