def call() {
    sh """
        helm upgrade --install argo-rollouts gitops/charts/argo-rollouts \
            --namespace argo-rollouts \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^ARGO_ROLLOUTS_/d' infra.env || true"
    sh "echo 'ARGO_ROLLOUTS_URL=http://argo-rollouts-argo-rollouts.argo-rollouts.svc.cluster.local:3100' >> infra.env"
    echo 'argo-rollouts installed'
}
return this
