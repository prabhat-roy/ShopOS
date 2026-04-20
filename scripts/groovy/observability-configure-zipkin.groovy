def call() {
    sh '''
        echo "=== Configure Zipkin ==="

        kubectl rollout status deploy/zipkin -n zipkin --timeout=120s || true

        ZIPKIN_IP=$(kubectl get svc zipkin -n zipkin \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "zipkin.zipkin.svc.cluster.local")

        sed -i '/^ZIPKIN_URL=/d' infra.env
        echo "ZIPKIN_URL=http://${ZIPKIN_IP}:9411" >> infra.env

        STATUS=$(curl -sf -o /dev/null -w "%{http_code}" \
            "http://${ZIPKIN_IP}:9411/health" 2>/dev/null || echo "000")
        echo "  Zipkin health: HTTP ${STATUS}"

        echo "Zipkin URL written to infra.env."
    '''
}
return this
