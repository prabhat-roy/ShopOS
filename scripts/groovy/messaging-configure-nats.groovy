def call() {
    sh """
        echo "Waiting for NATS..."
        until kubectl exec -n nats deploy/nats-nats -- \
            nats --server nats://localhost:4222 server check > /dev/null 2>&1; do sleep 10; done

        # Create JetStream streams for real-time ShopOS use cases
        kubectl exec -n nats deploy/nats-nats -- \
            nats --server nats://localhost:4222 stream add CHAT \
            --subjects "chat.>" --storage file --replicas 1 \
            --retention limits --max-age 24h --defaults || true

        kubectl exec -n nats deploy/nats-nats -- \
            nats --server nats://localhost:4222 stream add NOTIFICATIONS \
            --subjects "notify.>" --storage file --replicas 1 \
            --retention limits --max-age 72h --defaults || true

        kubectl exec -n nats deploy/nats-nats -- \
            nats --server nats://localhost:4222 stream add PRESENCE \
            --subjects "presence.>" --storage memory --replicas 1 \
            --retention limits --max-age 5m --defaults || true
    """
    echo 'nats configured — CHAT, NOTIFICATIONS, PRESENCE JetStream streams created'
}
return this
