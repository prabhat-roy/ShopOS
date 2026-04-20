def call() {
    sh """
        helm upgrade --install kibana observability/kibana/charts             --namespace kibana             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^KIBANA_/d' infra.env || true"
    sh "echo 'KIBANA_URL=http://kibana-kibana.kibana.svc.cluster.local:5601' >> infra.env"
    echo 'kibana installed'
}
return this
