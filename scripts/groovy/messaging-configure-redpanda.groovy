def call() {
    sh """
        echo "Waiting for Redpanda StatefulSet pod redpanda-0..."
        kubectl rollout status statefulset/redpanda -n redpanda --timeout=10m
        until kubectl exec -n redpanda pod/redpanda-0 -- \
            rpk cluster info > /dev/null 2>&1; do sleep 10; done

        # Create default topics
        kubectl exec -n redpanda pod/redpanda-0 -- \
            rpk topic create shopos-events --partitions 6 --replicas 1 || true
        kubectl exec -n redpanda pod/redpanda-0 -- \
            rpk topic create shopos-dlq --partitions 3 --replicas 1 || true

        echo "Redpanda topics created."
    """
    echo 'redpanda configured — shopos-events and shopos-dlq topics created'
}
return this
