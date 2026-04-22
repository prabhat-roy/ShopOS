def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install vcluster gitops/charts/vcluster \
            --namespace vcluster \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^VCLUSTER_/d' infra.env || true"
    sh "sed -i '/^VCLUSTER_URL=/d' infra.env 2>/dev/null || true; echo 'VCLUSTER_URL=https://vcluster-vcluster.vcluster.svc.cluster.local:8443' >> infra.env" 
    echo 'vcluster installed'
}
return this
