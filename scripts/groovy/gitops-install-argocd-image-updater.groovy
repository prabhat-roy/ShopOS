def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install argocd-image-updater gitops/charts/argocd-image-updater \
            --namespace argocd-image-updater \
            --create-namespace \
            --set config.argocd.serverAddress=http://argocd-server.argocd.svc.cluster.local \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^ARGOCD_IMAGE_UPDATER_/d' infra.env || true"
    sh "sed -i '/^ARGOCD_IMAGE_UPDATER_URL=/d' infra.env 2>/dev/null || true; echo 'ARGOCD_IMAGE_UPDATER_URL=http://argocd-image-updater.argocd-image-updater.svc.cluster.local:8080' >> infra.env" 
    echo 'argocd-image-updater installed'
}
return this
