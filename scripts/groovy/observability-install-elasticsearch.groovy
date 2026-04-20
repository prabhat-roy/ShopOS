def call() {
    sh """
        helm upgrade --install elasticsearch observability/elasticsearch/charts \
            --namespace elasticsearch --create-namespace --wait --timeout 10m
        ES_URL=http://elasticsearch-elasticsearch.elasticsearch.svc.cluster.local:9200
        sed -i '/^ELASTICSEARCH_URL=/d' infra.env || true
        echo "ELASTICSEARCH_URL=\${ES_URL}" >> infra.env
    """
    echo 'elasticsearch installed'
}
return this
