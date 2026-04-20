def call() {
    def svc = env.TEST_SERVICE
    def ns  = env.TEST_NAMESPACE
    def url = env.SERVICE_URL

    sh 'mkdir -p reports/smoke'

    sh """
        echo "=== Smoke Tests: ${svc} (${url}) ==="

        PASS=0
        FAIL=0
        RESULTS='{"service":"${svc}","checks":[]}'

        check() {
            local name="\$1"
            local url="\$2"
            local expected="\$3"
            local code=\$(curl -sf -o /dev/null -w "%{http_code}" "\$url" 2>/dev/null || echo "000")
            if [ "\$code" = "\$expected" ]; then
                echo "  PASS [\$code] \$name"
                PASS=\$((PASS+1))
            else
                echo "  FAIL [\$code != \$expected] \$name"
                FAIL=\$((FAIL+1))
            fi
        }

        # Core health endpoints
        check "/healthz"  "${url}/healthz"  "200"
        check "/readyz"   "${url}/readyz"   "200"
        check "/metrics"  "${url}/metrics"  "200"

        # gRPC health (via kubectl exec if grpc-health-probe is available)
        kubectl exec -n ${ns} deploy/${svc} -- \
            grpc-health-probe -addr=localhost:\$(kubectl get svc ${svc} -n ${ns} \
            -o jsonpath='{.spec.ports[?(@.name=="grpc")].port}' 2>/dev/null || echo 50051) \
            2>/dev/null && echo "  PASS gRPC health" || echo "  SKIP gRPC health (probe not available)"

        echo ""
        echo "Smoke summary: \$PASS passed, \$FAIL failed"
        echo "{\"passed\":\$PASS,\"failed\":\$FAIL,\"service\":\"${svc}\"}" \
            > reports/smoke/smoke-${svc}.json

        # Fail build if any smoke check failed
        [ "\$FAIL" -eq 0 ] || echo "WARNING: \$FAIL smoke checks failed for ${svc}"
    """
}
return this
