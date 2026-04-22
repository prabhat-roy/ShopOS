def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install fluxcd gitops/charts/fluxcd \
            --namespace flux-system \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^FLUXCD_/d' infra.env || true"
    sh "sed -i '/^FLUXCD_URL=/d' infra.env 2>/dev/null || true; echo 'FLUXCD_URL=http://fluxcd-fluxcd.flux-system.svc.cluster.local:9292' >> infra.env" 
    echo 'fluxcd installed'
}
return this
