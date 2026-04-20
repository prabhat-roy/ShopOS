def call() {
    sh 'mkdir -p reports/dast'

    def ns  = env.BUILD_DOMAIN
    def tag = env.IMAGE_TAG

    env.SERVICES.split(',').each { svc ->
        svc = svc.trim()

        // Resolve in-cluster service URL
        def targetUrl = "http://${svc}.${ns}.svc.cluster.local"

        // ── OWASP ZAP — passive + active scan via K8s API ─────────────────────
        if (env.ZAP_URL?.trim()) {
            sh """
                echo "=== DAST: OWASP ZAP — ${svc} ==="

                # Create new scan session
                ZAP_SESSION=\$(curl -sf "${env.ZAP_URL}/JSON/core/action/newSession/" \
                    -G --data-urlencode "name=${svc}-${tag}" | grep -o '"Result":"[^"]*"' | cut -d: -f2 | tr -d '"') || true

                # Spider the target
                SCAN_ID=\$(curl -sf "${env.ZAP_URL}/JSON/spider/action/scan/" \
                    -G --data-urlencode "url=${targetUrl}" \
                    --data-urlencode "recurse=true" \
                    | grep -o '"scan":"[^"]*"' | cut -d: -f2 | tr -d '"') || true

                # Wait for spider to finish
                if [ -n "\$SCAN_ID" ]; then
                    for i in \$(seq 1 30); do
                        STATUS=\$(curl -sf "${env.ZAP_URL}/JSON/spider/view/status/?scanId=\$SCAN_ID" \
                            | grep -o '"status":"[^"]*"' | cut -d: -f2 | tr -d '"') || true
                        [ "\$STATUS" = "100" ] && break
                        sleep 10
                    done
                fi

                # Active scan
                ASCAN_ID=\$(curl -sf "${env.ZAP_URL}/JSON/ascan/action/scan/" \
                    -G --data-urlencode "url=${targetUrl}" \
                    --data-urlencode "recurse=true" \
                    | grep -o '"scan":"[^"]*"' | cut -d: -f2 | tr -d '"') || true

                # Wait for active scan
                if [ -n "\$ASCAN_ID" ]; then
                    for i in \$(seq 1 60); do
                        STATUS=\$(curl -sf "${env.ZAP_URL}/JSON/ascan/view/status/?scanId=\$ASCAN_ID" \
                            | grep -o '"status":"[^"]*"' | cut -d: -f2 | tr -d '"') || true
                        [ "\$STATUS" = "100" ] && break
                        sleep 15
                    done
                fi

                # Export JSON report
                curl -sf "${env.ZAP_URL}/JSON/core/view/alerts/?baseurl=${targetUrl}&start=0&count=10000" \
                    > reports/dast/zap-${svc}.json 2>&1 || true

                echo "ZAP scan complete for ${svc}"
            """
        } else {
            echo "ZAP_URL not set — skipping ZAP DAST for ${svc}"
        }

        // ── Nuclei — template-based vulnerability scan via K8s API ────────────
        if (env.NUCLEI_URL?.trim()) {
            sh """
                echo "=== DAST: Nuclei — ${svc} ==="
                curl -sf -X POST "${env.NUCLEI_URL}/api/scan" \
                    -H "Content-Type: application/json" \
                    -d "{\"target\":\"${targetUrl}\",\"templates\":[\"cves\",\"vulnerabilities\",\"misconfiguration\",\"exposure\"],\"output\":\"json\"}" \
                    > reports/dast/nuclei-${svc}.json 2>&1 || true
            """
        } else {
            // Fallback: run Nuclei as docker container
            sh """
                echo "=== DAST: Nuclei (docker) — ${svc} ==="
                docker run --rm \
                    --network host \
                    projectdiscovery/nuclei:latest \
                    -target ${targetUrl} \
                    -t cves,vulnerabilities,misconfiguration,exposure \
                    -json \
                    -o /tmp/nuclei-${svc}.json 2>/dev/null || true
                cp /tmp/nuclei-${svc}.json reports/dast/nuclei-${svc}.json 2>/dev/null || true
            """
        }
    }

    echo 'DAST complete — reports/dast/'
}
return this
