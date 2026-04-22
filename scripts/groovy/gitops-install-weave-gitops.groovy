def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install weave-gitops gitops/charts/weave-gitops \
            --namespace weave-gitops \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^WEAVE_GITOPS_/d' infra.env || true"
    sh "sed -i '/^WEAVE_GITOPS_URL=/d' infra.env 2>/dev/null || true; echo 'WEAVE_GITOPS_URL=http://weave-gitops-weave-gitops.weave-gitops.svc.cluster.local:9001' >> infra.env" 
    sh "sed -i '/^WEAVE_GITOPS_USER=/d' infra.env 2>/dev/null || true; echo 'WEAVE_GITOPS_USER=admin' >> infra.env" 
    sh "sed -i '/^WEAVE_GITOPS_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'WEAVE_GITOPS_PASSWORD=admin' >> infra.env" 
    echo 'weave-gitops installed'
}
return this
