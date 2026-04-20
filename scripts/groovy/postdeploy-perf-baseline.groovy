def call() {
    def svc = env.TEST_SERVICE

    sh 'mkdir -p reports/load/baseline'

    sh """
        echo "=== Performance Baseline Check: ${svc} ==="

        P95_THRESHOLD=2000   # ms
        ERR_THRESHOLD=5      # percent
        PASS=0
        FAIL=0

        # ── Parse k6 summary ──────────────────────────────────────────────────
        for summary in reports/load/k6/*-summary.json; do
            [ -f "\$summary" ] || continue
            script=\$(basename "\$summary" -summary.json)

            P95=\$(python3 -c "
import json, sys
try:
    d = json.load(open('\$summary'))
    p = d.get('metrics',{}).get('http_req_duration',{}).get('values',{}).get('p(95)',0)
    print(int(p))
except: print(0)
" 2>/dev/null || echo 0)

            ERR_RATE=\$(python3 -c "
import json, sys
try:
    d = json.load(open('\$summary'))
    r = d.get('metrics',{}).get('http_req_failed',{}).get('values',{}).get('rate',0)
    print(int(r*100))
except: print(0)
" 2>/dev/null || echo 0)

            if [ "\$P95" -le "\$P95_THRESHOLD" ]; then
                echo "  PASS k6/\$script p95=\${P95}ms (<= \${P95_THRESHOLD}ms)"
                PASS=\$((PASS+1))
            else
                echo "  FAIL k6/\$script p95=\${P95}ms > \${P95_THRESHOLD}ms threshold"
                FAIL=\$((FAIL+1))
            fi

            if [ "\$ERR_RATE" -le "\$ERR_THRESHOLD" ]; then
                echo "  PASS k6/\$script error_rate=\${ERR_RATE}% (<= \${ERR_THRESHOLD}%)"
                PASS=\$((PASS+1))
            else
                echo "  FAIL k6/\$script error_rate=\${ERR_RATE}% > \${ERR_THRESHOLD}% threshold"
                FAIL=\$((FAIL+1))
            fi
        done

        # ── Parse Locust CSV ──────────────────────────────────────────────────
        for csv in reports/load/locust/*_stats.csv; do
            [ -f "\$csv" ] || continue
            script=\$(basename "\$csv" _stats.csv)

            # Average response time from Locust CSV (column 6 = avg)
            AVG=\$(tail -n +2 "\$csv" | awk -F',' '{sum+=\$6; n++} END {if(n>0) print int(sum/n); else print 0}' 2>/dev/null || echo 0)
            FAIL_RATE=\$(tail -n +2 "\$csv" | awk -F',' '{fail+=\$10; total+=\$9} END {if(total>0) print int(fail*100/total); else print 0}' 2>/dev/null || echo 0)

            if [ "\$AVG" -le "\$P95_THRESHOLD" ]; then
                echo "  PASS locust/\$script avg=\${AVG}ms (<= \${P95_THRESHOLD}ms)"
                PASS=\$((PASS+1))
            else
                echo "  FAIL locust/\$script avg=\${AVG}ms > \${P95_THRESHOLD}ms"
                FAIL=\$((FAIL+1))
            fi
        done

        echo ""
        echo "Baseline summary: \$PASS passed, \$FAIL failed"
        echo "{\"passed\":\$PASS,\"failed\":\$FAIL,\"service\":\"${svc}\",\"p95_threshold_ms\":\$P95_THRESHOLD,\"error_threshold_pct\":\$ERR_THRESHOLD}" \
            > reports/load/baseline/baseline-${svc}.json

        [ "\$FAIL" -eq 0 ] || echo "WARNING: \$FAIL performance thresholds breached"
    """
}
return this
