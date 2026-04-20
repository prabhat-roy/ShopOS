def call() {
    sh 'echo "=== Security Report Upload ==="'

    def services = env.SERVICES.split(',')

    // ── Helper: upload one file to DefectDojo ─────────────────────────────────
    def ddUpload = { String file, String scanType, String engagement ->
        sh """
            if [ -f "${file}" ] && [ -s "${file}" ]; then
                curl -sf -X POST "${env.DEFECTDOJO_URL}/api/v2/import-scan/" \
                    -H "Authorization: Token ${env.DEFECTDOJO_TOKEN}" \
                    -F "scan_type=${scanType}" \
                    -F "file=@${file}" \
                    -F "product_name=ShopOS" \
                    -F "engagement_name=${engagement}" \
                    -F "auto_create_context=true" \
                    -F "close_old_findings=false" || true
            fi
        """
    }

    // ── DefectDojo — upload all scan results ─────────────────────────────────
    if (env.DEFECTDOJO_TOKEN?.trim() && env.DEFECTDOJO_URL?.trim()) {
        services.each { svc ->
            svc = svc.trim()
            def eng    = "${svc}-ci"
            def engImg = "${svc}-ci-image"
            def engIac = "${svc}-ci-iac"
            def engDst = "${svc}-ci-dast"

            // SAST
            ddUpload('reports/sast/semgrep.json',                'Semgrep JSON Report',        eng)
            ddUpload('reports/sast/gosec.json',                  'Gosec Scanner',              eng)
            ddUpload('reports/sast/bandit.json',                 'Bandit Scan',                eng)
            ddUpload('reports/sast/pylint.json',                 'PyLint Scan',                eng)
            ddUpload('reports/sast/eslint.json',                 'ESLint Scan',                eng)
            ddUpload('reports/sast/golangci.json',               'GitLab SAST Report',         eng)
            ddUpload("reports/sast/spotbugs-${svc}.xml",         'SpotBugs Scan',              eng)
            ddUpload('reports/sast/pmd.json',                    'PMD Scan',                   eng)
            ddUpload('reports/sast/snyk-code.json',              'Snyk Scan',                  eng)

            // Secrets
            ddUpload('reports/secrets/gitleaks.json',            'Gitleaks Scan',              eng)
            ddUpload('reports/secrets/trufflehog.json',          'Trufflehog Scan',            eng)
            ddUpload('reports/secrets/detect-secrets.json',      'Detect-secrets Scan',        eng)

            // SCA / SBOM
            ddUpload('reports/sca/trivy-fs.json',                'Trivy Scan',                 eng)
            ddUpload('reports/sca/grype-fs.json',                'Anchore Grype',              eng)
            ddUpload('reports/sca/dependency-check.json',        'Dependency Check Scan',      eng)
            ddUpload('reports/sca/snyk-oss.json',                'Snyk Scan',                  eng)
            ddUpload("reports/sca/npm-audit-${svc}.json",        'NPM Audit Scan',             eng)
            ddUpload("reports/sca/govulncheck-${svc}.json",      'Go Vulnerability Check',     eng)

            // IaC
            ddUpload('reports/iac/results_json.json',            'Checkov Scan',               engIac)
            ddUpload('reports/iac/kics-results.json',            'KICS Scan',                  engIac)
            ddUpload('reports/iac/tfsec.json',                   'Tfsec Scan',                 engIac)
            ddUpload('reports/iac/terrascan-terraform.json',     'Terrascan Scan',             engIac)
            ddUpload('reports/iac/kubeaudit.json',               'Kubeaudit Scan',             engIac)

            // Image scans
            ddUpload("reports/image-scan/trivy-${svc}.json",     'Trivy Scan',                 engImg)
            ddUpload("reports/image-scan/grype-${svc}.json",     'Anchore Grype',              engImg)
            ddUpload("reports/image-scan/anchore-${svc}.json",   'Anchore Engine Scan',        engImg)

            // K8s Audit
            ddUpload('reports/k8s-audit/kube-bench.json',        'Kube Bench Scan',            engIac)
            ddUpload('reports/k8s-audit/kubescape.json',         'Kubescape Scan',             engIac)

            // DAST
            ddUpload("reports/dast/zap-${svc}.json",             'ZAP Scan',                   engDst)
            ddUpload("reports/dast/nuclei-${svc}.json",          'Nuclei Scan',                engDst)

            echo "DefectDojo upload complete for ${svc}"
        }
    } else {
        echo 'DEFECTDOJO_TOKEN not set — skipping DefectDojo upload'
    }

    // ── Dependency Track — SBOM upload ────────────────────────────────────────
    if (env.DEPTRACK_KEY?.trim() && env.DEPTRACK_URL?.trim()) {
        // Source-level SBOM
        sh """
            if [ -f reports/sca/sbom-cyclonedx.json ] && [ -s reports/sca/sbom-cyclonedx.json ]; then
                ENCODED=\$(base64 -w0 reports/sca/sbom-cyclonedx.json)
                curl -sf -X PUT "${env.DEPTRACK_URL}/api/v1/bom" \
                    -H "X-Api-Key: ${env.DEPTRACK_KEY}" \
                    -H "Content-Type: application/json" \
                    -d "{\"projectName\":\"ShopOS-${env.BUILD_DOMAIN}\",\"projectVersion\":\"${env.IMAGE_TAG}\",\"autoCreate\":true,\"bom\":\"\$ENCODED\"}" || true
                echo "Source SBOM uploaded to Dependency Track"
            fi
        """
        // Image SBOMs per service
        services.each { svc ->
            svc = svc.trim()
            sh """
                if [ -f reports/image-scan/sbom-image-${svc}.json ] && [ -s reports/image-scan/sbom-image-${svc}.json ]; then
                    ENCODED=\$(base64 -w0 reports/image-scan/sbom-image-${svc}.json)
                    curl -sf -X PUT "${env.DEPTRACK_URL}/api/v1/bom" \
                        -H "X-Api-Key: ${env.DEPTRACK_KEY}" \
                        -H "Content-Type: application/json" \
                        -d "{\"projectName\":\"${svc}\",\"projectVersion\":\"${env.IMAGE_TAG}\",\"autoCreate\":true,\"bom\":\"\$ENCODED\"}" || true
                fi
            """
        }
    } else {
        echo 'DEPTRACK_KEY not set — skipping Dependency Track upload'
    }

    // ── Wazuh — ship security events to SIEM ─────────────────────────────────
    if (env.WAZUH_URL?.trim() && env.WAZUH_TOKEN?.trim()) {
        sh """
            echo "=== Report: Wazuh SIEM event ==="
            curl -sf -X POST "${env.WAZUH_URL}/security/events" \
                -H "Authorization: Bearer ${env.WAZUH_TOKEN}" \
                -H "Content-Type: application/json" \
                -d "{\"events\":[{\"type\":\"pipeline\",\"service\":\"${env.SERVICES}\",\"tag\":\"${env.IMAGE_TAG}\",\"environment\":\"${env.BUILD_ENV}\",\"status\":\"scan_complete\",\"timestamp\":\"\$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}]}" || true
        """
    }

    echo 'Security report upload complete'
}
return this
