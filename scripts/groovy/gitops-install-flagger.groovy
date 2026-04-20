def call() {
    sh """
        helm upgrade --install flagger gitops/charts/flagger \
            --namespace flagger \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^FLAGGER_/d' infra.env || true"
    sh "echo 'FLAGGER_URL=http://flagger-flagger.flagger.svc.cluster.local:10080' >> infra.env"
    echo 'flagger installed'
}
return this
