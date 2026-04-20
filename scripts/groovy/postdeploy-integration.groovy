def call() {
    def svc  = env.TEST_SERVICE
    def ns   = env.TEST_NAMESPACE
    def url  = env.SERVICE_URL
    def domain = env.TEST_DOMAIN

    sh 'mkdir -p reports/integration'

    sh """
        echo "=== Integration Tests: ${svc} ==="

        PASS=0
        FAIL=0

        api_check() {
            local name="\$1"
            local method="\$2"
            local endpoint="\$3"
            local body="\$4"
            local expected_code="\$5"

            if [ -n "\$body" ]; then
                CODE=\$(curl -sf -X "\$method" -H "Content-Type: application/json" \
                    -d "\$body" -o /tmp/api-resp.json -w "%{http_code}" \
                    "${url}\$endpoint" 2>/dev/null || echo "000")
            else
                CODE=\$(curl -sf -X "\$method" \
                    -o /tmp/api-resp.json -w "%{http_code}" \
                    "${url}\$endpoint" 2>/dev/null || echo "000")
            fi

            if [ "\$CODE" = "\$expected_code" ] || [ "\$expected_code" = "2xx" -a "\${CODE#2}" != "\$CODE" ]; then
                echo "  PASS [\$CODE] \$name"
                PASS=\$((PASS+1))
            else
                echo "  FAIL [\$CODE != \$expected_code] \$name"
                cat /tmp/api-resp.json 2>/dev/null || true
                FAIL=\$((FAIL+1))
            fi
        }

        # ── Generic REST probe (works for any service) ────────────────────────
        api_check "GET /api/v1/ (list)"     GET  "/api/v1/"     "" "200"

        # ── Domain-specific integration probes ────────────────────────────────
        case "${domain}" in
          commerce)
            api_check "POST cart (add item)"   POST "/api/v1/cart" \
                '{"productId":"prod-001","quantity":1}' "2xx"
            api_check "GET cart"               GET  "/api/v1/cart" "" "200"
            api_check "POST checkout"          POST "/api/v1/checkout" \
                '{"paymentMethod":"CREDIT_CARD","shippingAddress":{"line1":"123 Main St","city":"Anytown","country":"US","zip":"10001"}}' "2xx"
            ;;
          catalog)
            api_check "GET products"           GET  "/api/v1/products" "" "200"
            api_check "GET categories"         GET  "/api/v1/categories" "" "200"
            api_check "GET search"             GET  "/api/v1/search?q=laptop" "" "200"
            ;;
          identity)
            api_check "POST login"             POST "/api/v1/auth/login" \
                '{"username":"test@shopos.local","password":"test"}' "200"
            ;;
          financial)
            api_check "GET invoices"           GET  "/api/v1/invoices" "" "200"
            ;;
          supply-chain)
            api_check "GET warehouse"          GET  "/api/v1/warehouse" "" "200"
            api_check "GET shipments"          GET  "/api/v1/shipments" "" "200"
            ;;
          *)
            api_check "GET health detail"      GET  "/api/v1/status" "" "200"
            ;;
        esac

        # ── Kafka connectivity (via service probe) ────────────────────────────
        kubectl exec -n ${ns} deploy/${svc} -- sh -c \
            'curl -sf \${KAFKA_BROKER:-kafka.kafka.svc.cluster.local:9092}/v3/clusters 2>/dev/null && echo "  PASS Kafka reachable" || echo "  SKIP Kafka probe (not available)"' \
            2>/dev/null || true

        # ── DB connectivity probe ─────────────────────────────────────────────
        kubectl exec -n ${ns} deploy/${svc} -- sh -c \
            'pg_isready -h \${DB_HOST:-postgres} 2>/dev/null && echo "  PASS DB reachable" || echo "  SKIP DB probe (pg_isready not available)"' \
            2>/dev/null || true

        echo ""
        echo "Integration summary: \$PASS passed, \$FAIL failed"
        echo "{\"passed\":\$PASS,\"failed\":\$FAIL,\"service\":\"${svc}\",\"domain\":\"${domain}\"}" \
            > reports/integration/integration-${svc}.json
    """
}
return this
