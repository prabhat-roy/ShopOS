def call() {
    sh """
        echo "Waiting for Redpanda..."
        until kubectl exec -n redpanda deploy/redpanda-redpanda -- \
            rpk cluster info > /dev/null 2>&1; do sleep 10; done

        # Create default topics
        kubectl exec -n redpanda deploy/redpanda-redpanda -- \
            rpk topic create shopos-events --partitions 6 --replicas 1 || true
        kubectl exec -n redpanda deploy/redpanda-redpanda -- \
            rpk topic create shopos-dlq --partitions 3 --replicas 1 || true

        echo "Redpanda topics created."
    """
    echo 'redpanda configured — shopos-events and shopos-dlq topics created'
}
return this
