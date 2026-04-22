def call() {
    sh """
        helm upgrade --install redpanda messaging/redpanda/charts \
            --namespace redpanda \
            --create-namespace \
            --set fullnameOverride=redpanda \
            --wait --timeout 5m
    """
    sh "sed -i '/^REDPANDA_/d' infra.env || true"
    sh "echo 'REDPANDA_URL=redpanda.redpanda.svc.cluster.local:9092' >> infra.env"
    sh "echo 'REDPANDA_ADMIN_URL=http://redpanda.redpanda.svc.cluster.local:9644' >> infra.env"
    echo 'redpanda installed'
}
return this
