def call() {
    sh """
        helm upgrade --install nats messaging/nats/charts \
            --namespace nats \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^NATS_/d' infra.env || true"
    sh "sed -i '/^NATS_URL=/d' infra.env 2>/dev/null || true; echo 'NATS_URL=nats://nats-nats.nats.svc.cluster.local:4222' >> infra.env" 
    sh "sed -i '/^NATS_MONITORING_URL=/d' infra.env 2>/dev/null || true; echo 'NATS_MONITORING_URL=http://nats-nats.nats.svc.cluster.local:8222' >> infra.env" 
    echo 'nats installed'
}
return this
