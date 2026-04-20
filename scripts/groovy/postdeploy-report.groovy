def call() {
    def svc     = env.TEST_SERVICE
    def buildNo = env.BUILD_NUMBER
    def tag     = env.IMAGE_TAG ?: 'unknown'

    sh """
        echo "=== Post-Deploy Report: ${svc} (build ${buildNo}) ==="
        mkdir -p reports/summary
    """

    // ── Aggregate all JSON reports into one summary ────────────────────────────
    sh """
        python3 - <<'PYEOF'
import json, os, glob

summary = {
    "service":    "${svc}",
    "build":      "${buildNo}",
    "image_tag":  "${tag}",
    "sections":   {}
}

def load_json(path):
    try:
        with open(path) as f:
            return json.load(f)
    except:
        return {}

# Smoke
for f in glob.glob("reports/smoke/*.json"):
    summary["sections"]["smoke"] = load_json(f)

# Integration
for f in glob.glob("reports/integration/*.json"):
    summary["sections"]["integration"] = load_json(f)

# Baseline
for f in glob.glob("reports/load/baseline/*.json"):
    summary["sections"]["performance_baseline"] = load_json(f)

# SLO
for f in glob.glob("reports/slo/*.json"):
    summary["sections"]["slo"] = load_json(f)

# k6 summaries
k6 = {}
for f in glob.glob("reports/load/k6/*-summary.json"):
    name = os.path.basename(f).replace("-summary.json", "")
    d = load_json(f)
    k6[name] = {
        "p95_ms":     d.get("metrics", {}).get("http_req_duration", {}).get("values", {}).get("p(95)", 0),
        "error_rate": d.get("metrics", {}).get("http_req_failed",    {}).get("values", {}).get("rate",  0),
        "rps":        d.get("metrics", {}).get("http_reqs",          {}).get("values", {}).get("rate",  0),
    }
if k6:
    summary["sections"]["k6"] = k6

# Chaos
chaos = []
for f in glob.glob("reports/chaos/*.json"):
    d = load_json(f)
    if isinstance(d, dict):
        chaos.append(d)
if chaos:
    summary["sections"]["chaos"] = chaos

# Overall pass/fail
total_fail = sum(
    s.get("failed", 0)
    for s in summary["sections"].values()
    if isinstance(s, dict)
)
summary["overall_failures"] = total_fail
summary["result"] = "PASS" if total_fail == 0 else "FAIL"

out = "reports/summary/post-deploy-${svc}-${buildNo}.json"
with open(out, "w") as f:
    json.dump(summary, f, indent=2)

print(f"Summary written to {out}")
print(f"Overall result: {summary['result']} ({total_fail} failures)")
PYEOF
    """

    // ── Upload k6 results to InfluxDB (for Grafana k6 dashboard) ─────────────
    if (env.INFLUXDB_URL?.trim()) {
        sh """
            for f in reports/load/k6/*.json; do
                [ -f "\$f" ] || continue
                curl -sf -X POST "${env.INFLUXDB_URL}/write?db=k6" \
                    --data-binary @"\$f" || true
            done
            echo "k6 results uploaded to InfluxDB"
        """
    }

    // ── Upload Locust CSV to reporting system (optional) ──────────────────────
    if (env.DEFECTDOJO_URL?.trim() && env.DEFECTDOJO_TOKEN?.trim()) {
        sh """
            # Upload SLO report as generic finding
            if [ -f reports/slo/slo-${svc}.json ]; then
                curl -sf -X POST "${env.DEFECTDOJO_URL}/api/v2/import-scan/" \
                    -H "Authorization: Token ${env.DEFECTDOJO_TOKEN}" \
                    -F "scan_type=Generic Findings Import" \
                    -F "file=@reports/slo/slo-${svc}.json" \
                    -F "product_name=ShopOS" \
                    -F "engagement_name=${svc}-post-deploy" \
                    -F "auto_create_context=true" || true
            fi
        """
    }

    // ── Print human-readable summary to console ───────────────────────────────
    sh """
        echo ""
        echo "╔══════════════════════════════════════════════════════╗"
        echo "║          POST-DEPLOY PIPELINE SUMMARY                ║"
        echo "╠══════════════════════════════════════════════════════╣"
        python3 -c "
import json, glob
for f in glob.glob('reports/summary/*.json'):
    d = json.load(open(f))
    print(f'  Service   : {d[\"service\"]}')
    print(f'  Build     : {d[\"build\"]}')
    print(f'  Image tag : {d[\"image_tag\"]}')
    print(f'  Result    : {d[\"result\"]}')
    print()
    for section, data in d.get('sections', {}).items():
        if isinstance(data, dict):
            p = data.get('passed', '-')
            f2 = data.get('failed', '-')
            print(f'  {section:<25} passed={p}  failed={f2}')
        elif isinstance(data, list):
            print(f'  {section:<25} {len(data)} experiment(s)')
" 2>/dev/null || cat reports/summary/*.json 2>/dev/null || true
        echo "╚══════════════════════════════════════════════════════╝"
    """
}
return this
