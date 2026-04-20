def call() {
    sh """
        helm upgrade --install tempo observability/tempo/charts             --namespace tempo             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^TEMPO_/d' infra.env || true"
    sh "echo 'TEMPO_URL=http://tempo-tempo.tempo.svc.cluster.local:3200' >> infra.env"
    echo 'tempo installed'
}
return this
