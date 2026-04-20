def call() {
    sh """
        helm upgrade --install nats messaging/nats/charts \
            --namespace nats \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^NATS_/d' infra.env || true"
    sh "echo 'NATS_URL=nats://nats-nats.nats.svc.cluster.local:4222' >> infra.env"
    sh "echo 'NATS_MONITORING_URL=http://nats-nats.nats.svc.cluster.local:8222' >> infra.env"
    echo 'nats installed'
}
return this
