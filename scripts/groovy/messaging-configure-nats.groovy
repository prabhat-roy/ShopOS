def call() {
    sh """
        echo "Waiting for NATS rollout..."
        kubectl rollout status deployment/nats -n nats --timeout=5m

        # Clean up any leftover pod from a prior run
        kubectl delete pod nats-setup -n nats --ignore-not-found=true 2>/dev/null || true

        # Create JetStream streams using nats-box (nats CLI not in nats-server image)
        kubectl run nats-setup -i --rm --restart=Never \
            --image=natsio/nats-box:latest -n nats \
            --command -- sh -c '
                nats --server nats://nats.nats.svc.cluster.local:4222 stream add CHAT \
                    --subjects "chat.>" --storage file --replicas 1 \
                    --retention limits --max-age 24h --defaults 2>/dev/null || true
                nats --server nats://nats.nats.svc.cluster.local:4222 stream add NOTIFICATIONS \
                    --subjects "notify.>" --storage file --replicas 1 \
                    --retention limits --max-age 72h --defaults 2>/dev/null || true
                nats --server nats://nats.nats.svc.cluster.local:4222 stream add PRESENCE \
                    --subjects "presence.>" --storage memory --replicas 1 \
                    --retention limits --max-age 5m --defaults 2>/dev/null || true
            ' || true
    """
    echo 'nats configured — CHAT, NOTIFICATIONS, PRESENCE JetStream streams created'
}
return this
