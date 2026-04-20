def call() {
    sh '''
        echo "=== Configure Argo Rollouts ==="

        kubectl rollout status deploy/argo-rollouts -n argo-rollouts --timeout=120s || true

        # Create a default AnalysisTemplate based on Prometheus success rate
        PROM_URL=$(grep '^PROMETHEUS_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://prometheus-prometheus.prometheus.svc.cluster.local:9090")

        kubectl apply -f - <<EOF
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: success-rate
  namespace: argo-rollouts
spec:
  args:
  - name: service-name
  metrics:
  - name: success-rate
    interval: 30s
    successCondition: result[0] >= 0.95
    failureLimit: 3
    provider:
      prometheus:
        address: ${PROM_URL}
        query: |
          sum(rate(http_requests_total{service="{{args.service-name}}",status!~"5.."}[2m]))
          /
          sum(rate(http_requests_total{service="{{args.service-name}}"}[2m]))
EOF

        echo "Argo Rollouts default success-rate AnalysisTemplate created."
    '''
}
return this
