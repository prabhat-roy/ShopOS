def call() {
    sh """
        helm upgrade --install goldilocks observability/goldilocks/charts \
            --namespace monitoring \
            --create-namespace \
            --set fullnameOverride=goldilocks \
            --wait --timeout 5m
    """
    sh "sed -i '/^GOLDILOCKS_/d' infra.env || true"
    sh "echo 'GOLDILOCKS_URL=http://goldilocks.monitoring.svc.cluster.local:80' >> infra.env"
    echo 'Goldilocks installed'
}
return this
