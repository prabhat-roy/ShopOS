def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install flagger gitops/charts/flagger \
            --namespace flagger \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^FLAGGER_/d' infra.env || true"
    sh "sed -i '/^FLAGGER_URL=/d' infra.env 2>/dev/null || true; echo 'FLAGGER_URL=http://flagger-flagger.flagger.svc.cluster.local:10080' >> infra.env" 
    echo 'flagger installed'
}
return this
