def call() {
    sh """
        helm upgrade --install fluent-bit observability/fluent-bit/charts             --namespace fluent-bit             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^FLUENT_BIT_/d' infra.env || true"
    sh "sed -i '/^FLUENT_BIT_URL=/d' infra.env 2>/dev/null || true; echo 'FLUENT_BIT_URL=http://fluent-bit-fluent-bit.fluent-bit.svc.cluster.local:2020' >> infra.env" 
    echo 'fluent-bit installed'
}
return this
