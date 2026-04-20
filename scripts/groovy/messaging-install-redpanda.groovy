def call() {
    sh """
        helm upgrade --install redpanda messaging/redpanda/charts \
            --namespace redpanda \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^REDPANDA_/d' infra.env || true"
    sh "sed -i '/^REDPANDA_URL=/d' infra.env 2>/dev/null || true; echo 'REDPANDA_URL=redpanda-redpanda.redpanda.svc.cluster.local:9092' >> infra.env" 
    sh "sed -i '/^REDPANDA_ADMIN_URL=/d' infra.env 2>/dev/null || true; echo 'REDPANDA_ADMIN_URL=http://redpanda-redpanda.redpanda.svc.cluster.local:9644' >> infra.env" 
    echo 'redpanda installed'
}
return this
