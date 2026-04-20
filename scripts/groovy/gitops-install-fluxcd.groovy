def call() {
    sh """
        helm upgrade --install fluxcd gitops/charts/fluxcd \
            --namespace flux-system \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^FLUXCD_/d' infra.env || true"
    sh "echo 'FLUXCD_URL=http://fluxcd-fluxcd.flux-system.svc.cluster.local:9292' >> infra.env"
    echo 'fluxcd installed'
}
return this
