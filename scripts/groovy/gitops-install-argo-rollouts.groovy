def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install argo-rollouts gitops/charts/argo-rollouts \
            --namespace argo-rollouts \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^ARGO_ROLLOUTS_/d' infra.env || true"
    sh "sed -i '/^ARGO_ROLLOUTS_URL=/d' infra.env 2>/dev/null || true; echo 'ARGO_ROLLOUTS_URL=http://argo-rollouts-argo-rollouts.argo-rollouts.svc.cluster.local:3100' >> infra.env" 
    echo 'argo-rollouts installed'
}
return this
