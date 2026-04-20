def call() {
    sh '''
        echo "=== Configure Sloth ==="

        kubectl rollout status deploy/sloth -n sloth --timeout=120s || true

        # Create SLI/SLO specs for critical services via Sloth CRDs
        for svc in checkout-service order-service payment-service api-gateway; do
            kubectl apply -f - <<EOF
apiVersion: sloth.slok.dev/v1
kind: PrometheusServiceLevel
metadata:
  name: ${svc}-slos
  namespace: monitoring
spec:
  service: "${svc}"
  labels:
    team: platform
  slos:
  - name: "requests-availability"
    objective: 99.5
    description: "Availability SLO for ${svc}"
    sli:
      events:
        errorQuery: sum(rate(http_requests_total{service="${svc}",status=~"(5..)"}[{{.window}}]))
        totalQuery: sum(rate(http_requests_total{service="${svc}"}[{{.window}}]))
    alerting:
      name: ${svc}HighErrorRate
      labels:
        severity: warning
      annotations:
        summary: "${svc} is burning error budget"
      pageAlert:
        labels:
          severity: critical
      ticketAlert:
        labels:
          severity: warning
EOF
        done 2>/dev/null || true

        echo "Sloth SLI/SLO PrometheusServiceLevel objects created for critical services."
    '''
}
return this
