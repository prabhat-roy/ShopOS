def call() {
    sh """
        helm upgrade --install zipkin observability/zipkin/charts             --namespace zipkin             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^ZIPKIN_/d' infra.env || true"
    sh "echo 'ZIPKIN_URL=http://zipkin-zipkin.zipkin.svc.cluster.local:9411' >> infra.env"
    echo 'zipkin installed'
}
return this
