def call() {
    def svc    = env.TEST_SERVICE
    def ns     = env.TEST_NAMESPACE
    def prom   = env.PROM_URL
    def grafana = env.GRAFANA_URL

    sh 'mkdir -p reports/slo'

    sh """
        echo "=== SLO & Observability Validation: ${svc} ==="

        PASS=0
        FAIL=0

        prom_query() {
            local query="\$1"
            curl -sf "${prom}/api/v1/query" \
                --data-urlencode "query=\$query" \
                2>/dev/null | python3 -c "
import json, sys
try:
    d = json.load(sys.stdin)
    r = d['data']['result']
    print(r[0]['value'][1] if r else '0')
except Exception as e:
    print('0')
" 2>/dev/null || echo "0"
        }

        # ── Availability SLO (99.5% over last 30m) ────────────────────────────
        AVAIL=\$(prom_query 'sum(rate(http_requests_total{service="${svc}",status!~"5.."}[30m])) / sum(rate(http_requests_total{service="${svc}"}[30m])) * 100')
        echo "  Availability (30m): \${AVAIL}%"
        python3 -c "v=float('\${AVAIL}' or 0); exit(0 if v>=99.5 else 1)" 2>/dev/null \
            && { echo "  PASS availability >= 99.5%"; PASS=\$((PASS+1)); } \
            || { echo "  FAIL availability \${AVAIL}% < 99.5%"; FAIL=\$((FAIL+1)); }

        # ── Latency SLO (p95 < 2s over last 30m) ─────────────────────────────
        P95=\$(prom_query 'histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service="${svc}"}[30m])) by (le)) * 1000')
        echo "  p95 latency (30m): \${P95}ms"
        python3 -c "v=float('\${P95}' or 0); exit(0 if v<2000 else 1)" 2>/dev/null \
            && { echo "  PASS p95 < 2000ms"; PASS=\$((PASS+1)); } \
            || { echo "  FAIL p95 \${P95}ms >= 2000ms"; FAIL=\$((FAIL+1)); }

        # ── Error budget remaining (via Pyrra if available) ────────────────────
        if [ -n "${env.PYRRA_URL}" ]; then
            BUDGET=\$(curl -sf "${env.PYRRA_URL}/api/v1/objectives" 2>/dev/null \
                | python3 -c "
import json,sys
try:
    d=json.load(sys.stdin)
    obj=[o for o in d if '${svc}' in str(o)]
    print(obj[0].get('errorBudgetRemaining','N/A') if obj else 'N/A')
except: print('N/A')
" 2>/dev/null || echo "N/A")
            echo "  Error budget remaining: \${BUDGET}"
        fi

        # ── Prometheus scraping the service ────────────────────────────────────
        SCRAPE_UP=\$(prom_query 'up{job="${svc}"}')
        echo "  Prometheus scrape status: \${SCRAPE_UP}"
        [ "\$SCRAPE_UP" = "1" ] \
            && { echo "  PASS Prometheus scraping ${svc}"; PASS=\$((PASS+1)); } \
            || { echo "  WARN Prometheus not scraping ${svc} (job may use different label)"; }

        # ── Active alerts for this service ────────────────────────────────────
        ALERTS=\$(curl -sf "${prom}/api/v1/alerts" 2>/dev/null \
            | python3 -c "
import json,sys
try:
    d=json.load(sys.stdin)
    firing=[a for a in d['data']['alerts'] if a['state']=='firing' and '${svc}' in str(a['labels'])]
    print(len(firing))
except: print(0)
" 2>/dev/null || echo 0)
        echo "  Active alerts for ${svc}: \${ALERTS}"
        [ "\$ALERTS" = "0" ] \
            && { echo "  PASS no firing alerts"; PASS=\$((PASS+1)); } \
            || { echo "  WARN \${ALERTS} alert(s) firing for ${svc}"; FAIL=\$((FAIL+1)); }

        # ── Grafana dashboard check ────────────────────────────────────────────
        GF_STATUS=\$(curl -sf -o /dev/null -w "%{http_code}" \
            "${grafana}/api/health" 2>/dev/null || echo "000")
        [ "\$GF_STATUS" = "200" ] \
            && { echo "  PASS Grafana healthy"; PASS=\$((PASS+1)); } \
            || { echo "  WARN Grafana returned \$GF_STATUS"; }

        # ── Check Loki receiving logs ──────────────────────────────────────────
        if [ -n "${env.LOKI_URL}" ]; then
            LOKI_STATUS=\$(curl -sf -o /dev/null -w "%{http_code}" \
                "${env.LOKI_URL}/ready" 2>/dev/null || echo "000")
            [ "\$LOKI_STATUS" = "200" ] \
                && echo "  PASS Loki ready" \
                || echo "  WARN Loki returned \$LOKI_STATUS"
        fi

        echo ""
        echo "SLO summary: \$PASS passed, \$FAIL failed"
        cat > reports/slo/slo-${svc}.json <<EOF
{
  "service":      "${svc}",
  "availability": "\${AVAIL}",
  "p95_ms":       "\${P95}",
  "alerts":       "\${ALERTS}",
  "passed":       \$PASS,
  "failed":       \$FAIL
}
EOF
    """
}
return this
