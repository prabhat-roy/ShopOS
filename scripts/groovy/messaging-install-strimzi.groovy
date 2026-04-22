def call() {
    sh """
        helm repo add strimzi https://strimzi.io/charts/ 2>/dev/null || true
        helm repo update strimzi
        helm upgrade --install strimzi strimzi/strimzi-kafka-operator \
            --namespace strimzi \
            --create-namespace \
            --version 0.51.0 \
            --set watchNamespaces={kafka} \
            --wait --timeout 5m
    """
    sh "sed -i '/^STRIMZI_/d' infra.env || true"
    sh "echo 'STRIMZI_NAMESPACE=kafka' >> infra.env"
    echo 'strimzi installed'
}
return this
