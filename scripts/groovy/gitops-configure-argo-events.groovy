def call() {
    sh '''
        echo "=== Configure Argo Events ==="

        kubectl rollout status deploy/controller-manager -n argo-events --timeout=120s || true

        # Create a native EventBus (NATS-based)
        kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: EventBus
metadata:
  name: default
  namespace: argo-events
spec:
  nats:
    native:
      replicas: 3
      auth: token
EOF

        echo "Argo Events default EventBus created."
    '''
}
return this
