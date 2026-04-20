def call() {
    sh """
        helm upgrade --install argo-workflows gitops/charts/argo-workflows \
            --namespace argo-workflows \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^ARGO_WORKFLOWS_/d' infra.env || true"
    sh "echo 'ARGO_WORKFLOWS_URL=http://argo-workflows-argo-workflows.argo-workflows.svc.cluster.local:2746' >> infra.env"
    sh "echo 'ARGO_WORKFLOWS_USER=admin' >> infra.env"
    sh "echo 'ARGO_WORKFLOWS_PASSWORD=admin' >> infra.env"
    echo 'argo-workflows installed'
}
return this
