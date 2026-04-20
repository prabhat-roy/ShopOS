def call() {
    sh """
        helm upgrade --install weave-gitops gitops/charts/weave-gitops \
            --namespace weave-gitops \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^WEAVE_GITOPS_/d' infra.env || true"
    sh "echo 'WEAVE_GITOPS_URL=http://weave-gitops-weave-gitops.weave-gitops.svc.cluster.local:9001' >> infra.env"
    sh "echo 'WEAVE_GITOPS_USER=admin' >> infra.env"
    sh "echo 'WEAVE_GITOPS_PASSWORD=admin' >> infra.env"
    echo 'weave-gitops installed'
}
return this
