def call() {
    sh """
        helm upgrade --install pulsar messaging/pulsar/charts \
            --namespace pulsar \
            --create-namespace \
            --set fullnameOverride=pulsar \
            --wait --timeout 10m
    """
    sh "sed -i '/^PULSAR_/d' infra.env || true"
    sh "echo 'PULSAR_URL=pulsar://pulsar.pulsar.svc.cluster.local:6650' >> infra.env"
    sh "echo 'PULSAR_HTTP_URL=http://pulsar.pulsar.svc.cluster.local:8080' >> infra.env"
    echo 'pulsar installed'
}
return this
