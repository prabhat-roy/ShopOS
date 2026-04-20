def call() {
    sh '''
        echo "=== Configure Flagger ==="

        kubectl rollout status deploy/flagger -n flagger --timeout=120s || true

        # Verify Flagger can talk to Prometheus
        PROM_URL=$(grep '^PROMETHEUS_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://prometheus-prometheus.prometheus.svc.cluster.local:9090")

        # Create a MetricTemplate for HTTP success rate (reusable across Canary resources)
        kubectl apply -f - <<EOF
apiVersion: flagger.app/v1beta1
kind: MetricTemplate
metadata:
  name: http-success-rate
  namespace: flagger
spec:
  provider:
    type: prometheus
    address: ${PROM_URL}
  query: |
    100 - sum(
      rate(http_requests_total{namespace="{{ namespace }}",service="{{ target }}",status=~"5.."}[{{ interval }}])
    ) /
    sum(
      rate(http_requests_total{namespace="{{ namespace }}",service="{{ target }}"}[{{ interval }}])
    ) * 100
EOF

        echo "Flagger HTTP success rate MetricTemplate created."
    '''
}
return this
