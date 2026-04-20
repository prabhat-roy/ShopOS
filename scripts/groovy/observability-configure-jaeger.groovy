def call() {
    sh '''
        echo "=== Configure Jaeger ==="

        kubectl rollout status deploy/jaeger -n jaeger --timeout=120s || true

        JAEGER_IP=$(kubectl get svc jaeger-query -n jaeger \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "jaeger-query.jaeger.svc.cluster.local")
        JAEGER_COLLECTOR=$(kubectl get svc jaeger-collector -n jaeger \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "jaeger-collector.jaeger.svc.cluster.local")

        sed -i '/^JAEGER_URL=/d; /^JAEGER_COLLECTOR_URL=/d' infra.env
        echo "JAEGER_URL=http://${JAEGER_IP}:16686" >> infra.env
        echo "JAEGER_COLLECTOR_URL=http://${JAEGER_COLLECTOR}:14268" >> infra.env

        # Verify query API
        STATUS=$(curl -sf -o /dev/null -w "%{http_code}" \
            "http://${JAEGER_IP}:16686/api/services" 2>/dev/null || echo "000")
        echo "  Jaeger query API: HTTP ${STATUS}"

        echo "Jaeger URLs written to infra.env."
    '''
}
return this
