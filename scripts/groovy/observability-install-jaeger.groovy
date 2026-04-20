def call() {
    sh """
        helm upgrade --install jaeger observability/jaeger/charts             --namespace jaeger             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^JAEGER_/d' infra.env || true"
    sh "echo 'JAEGER_URL=http://jaeger-jaeger.jaeger.svc.cluster.local:16686' >> infra.env"
    echo 'jaeger installed'
}
return this
