def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install thanos observability/thanos/charts \
            --namespace thanos \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^THANOS_/d' infra.env || true"
    sh "sed -i '/^THANOS_URL=/d' infra.env 2>/dev/null || true; echo 'THANOS_URL=http://thanos-thanos.thanos.svc.cluster.local:10902' >> infra.env"
    echo 'thanos installed'
}
return this
