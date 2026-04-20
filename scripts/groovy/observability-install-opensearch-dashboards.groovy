def call() {
    sh """
        helm upgrade --install opensearch-dashboards observability/opensearch-dashboards/charts             --namespace opensearch-dashboards             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^OPENSEARCH_DASHBOARDS_/d' infra.env || true"
    sh "sed -i '/^OPENSEARCH_DASHBOARDS_URL=/d' infra.env 2>/dev/null || true; echo 'OPENSEARCH_DASHBOARDS_URL=http://opensearch-dashboards-opensearch-dashboards.opensearch-dashboards.svc.cluster.local:5601' >> infra.env" 
    echo 'opensearch-dashboards installed'
}
return this
