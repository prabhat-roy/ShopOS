def call() {
    sh """
        helm upgrade --install node-exporter observability/node-exporter/charts             --namespace node-exporter             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^NODE_EXPORTER_/d' infra.env || true"
    sh "echo 'NODE_EXPORTER_URL=http://node-exporter-node-exporter.node-exporter.svc.cluster.local:9100' >> infra.env"
    echo 'node-exporter installed'
}
return this
