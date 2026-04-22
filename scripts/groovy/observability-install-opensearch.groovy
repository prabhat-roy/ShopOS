def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install opensearch observability/opensearch/charts \
            --namespace opensearch --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 10m
        OS_URL=http://opensearch-opensearch.opensearch.svc.cluster.local:9200
        sed -i '/^OPENSEARCH_URL=/d' infra.env || true
        echo "OPENSEARCH_URL=\${OS_URL}" >> infra.env
    """
    echo 'opensearch installed'
}
return this
