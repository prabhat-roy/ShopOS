def call() {
    sh '''
        echo "=== Configure Pyrra ==="

        kubectl rollout status deploy/pyrra -n pyrra --timeout=120s || true

        PYRRA_IP=$(kubectl get svc pyrra -n pyrra \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "pyrra.pyrra.svc.cluster.local")
        PYRRA_URL="http://${PYRRA_IP}:9099"

        sed -i '/^PYRRA_URL=/d' infra.env
        echo "PYRRA_URL=${PYRRA_URL}" >> infra.env

        # Create SLO objects for critical ShopOS services
        for svc in checkout-service order-service payment-service auth-service; do
            kubectl apply -f - <<EOF
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: ${svc}-availability
  namespace: monitoring
  labels:
    pyrra.dev/component: ${svc}
spec:
  target: "99.5"
  window: 4w
  description: "${svc} HTTP availability SLO"
  indicator:
    ratio:
      errors:
        metric: http_requests_total{service="${svc}",status=~"5.."}
      total:
        metric: http_requests_total{service="${svc}"}
EOF
        done 2>/dev/null || true

        echo "Pyrra URL written to infra.env. SLO objects created for critical services."
    '''
}
return this
