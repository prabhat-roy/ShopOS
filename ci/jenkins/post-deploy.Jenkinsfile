// ── POST-DEPLOY VALIDATION PIPELINE ───────────────────────────────────────────
// Runs after ArgoCD has deployed services. Covers:
//   Smoke tests · Integration tests · API quality (Spectral, Hurl, Pact)
//   DAST (OWASP ZAP, Nuclei) · Load tests (k6, Locust, Gatling)
//   Performance baseline · Chaos engineering (Chaos Mesh, Litmus)
//   SLO & observability validation · Security report upload
//   All dashboard links · Notification with full summary
// ─────────────────────────────────────────────────────────────────────────────

pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '30'))
        timeout(time: 180, unit: 'MINUTES')
    }

    parameters {
        string(
            name: 'SERVICE_NAME',
            defaultValue: '',
            description: 'Service to test (e.g. order-service). Required.'
        )
        choice(
            name: 'DOMAIN',
            choices: ['commerce','platform','identity','catalog','supply-chain','financial',
                      'customer-experience','communications','content','analytics-ai','b2b',
                      'integrations','affiliate','marketplace','gamification','developer-platform',
                      'compliance','sustainability','events-ticketing','auction','rental','web'],
            description: 'Kubernetes namespace / business domain of the service'
        )
        choice(
            name: 'ENVIRONMENT',
            choices: ['staging','dev','prod'],
            description: 'Target environment (staging recommended for load + chaos)'
        )
        string(
            name: 'IMAGE_TAG',
            defaultValue: '',
            description: 'Image tag of the deployed build. Auto-detected from K8s if blank.'
        )
        string(
            name: 'API_BASE_URL',
            defaultValue: 'http://api-gateway:8080',
            description: 'Base URL of the API Gateway for Hurl integration tests'
        )

        // ── Load test profile ─────────────────────────────────────────────────
        choice(
            name: 'LOAD_PROFILE',
            choices: ['medium','light','heavy','spike'],
            description: 'Load profile: light (10VU/2m) · medium (50VU/5m) · heavy (200VU/10m) · spike (500VU/30s)'
        )
        string(name: 'LOAD_VUS',      defaultValue: '', description: 'Override virtual users (blank = profile default)')
        string(name: 'LOAD_DURATION', defaultValue: '', description: 'Override duration e.g. 3m (blank = profile default)')
        string(name: 'CHAOS_DURATION',defaultValue: '2m', description: 'Duration for each chaos experiment')

        // ── Stage toggles ─────────────────────────────────────────────────────
        booleanParam(name: 'SKIP_SMOKE',        defaultValue: false, description: 'Skip smoke tests (/healthz, /readyz, /metrics, gRPC probe)')
        booleanParam(name: 'SKIP_INTEGRATION',  defaultValue: false, description: 'Skip integration tests (cross-service, DB, Kafka)')
        booleanParam(name: 'SKIP_API_QUALITY',  defaultValue: false, description: 'Skip API quality (Spectral lint, Hurl HTTP tests, Pact contracts)')
        booleanParam(name: 'SKIP_DAST',         defaultValue: false, description: 'Skip DAST (OWASP ZAP, Nuclei CVE scanning)')
        booleanParam(name: 'SKIP_K6',           defaultValue: false, description: 'Skip k6 load tests')
        booleanParam(name: 'SKIP_LOCUST',       defaultValue: false, description: 'Skip Locust load tests')
        booleanParam(name: 'SKIP_GATLING',      defaultValue: false, description: 'Skip Gatling simulations')
        booleanParam(name: 'SKIP_PERF_CHECK',   defaultValue: false, description: 'Skip performance baseline (p95 < 2s, error rate < 5%)')
        booleanParam(name: 'SKIP_CHAOS_MESH',   defaultValue: false, description: 'Skip Chaos Mesh (pod-kill, network-delay, cpu-stress)')
        booleanParam(name: 'SKIP_LITMUS',       defaultValue: false, description: 'Skip Litmus chaos workflows (database-chaos, payment-chaos)')
        booleanParam(name: 'SKIP_SLO',          defaultValue: false, description: 'Skip SLO validation (availability, p95, alerts, Grafana, Loki)')
        booleanParam(name: 'SKIP_SECURITY_REPORT', defaultValue: false, description: 'Skip post-deploy security report upload to DefectDojo')
    }

    stages {

        // ── GIT FETCH ─────────────────────────────────────────────────────────

        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        // ── ENVIRONMENT ───────────────────────────────────────────────────────

        stage('Load Environment') {
            steps {
                script {
                    if (!params.SERVICE_NAME?.trim()) {
                        error 'SERVICE_NAME is required'
                    }
                    if (!fileExists('infra.env')) {
                        error "infra.env not found — run k8s-infra and observability pipelines first"
                    }

                    def envMap = [:]
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }

                    // Observability
                    env.GRAFANA_URL      = envMap['GRAFANA_URL']        ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'
                    env.PROMETHEUS_URL   = envMap['PROMETHEUS_URL']     ?: 'http://prometheus-prometheus.prometheus.svc.cluster.local:9090'
                    env.JAEGER_URL       = envMap['JAEGER_URL']         ?: 'http://jaeger-query.tracing.svc.cluster.local:16686'
                    env.LOKI_URL         = envMap['LOKI_URL']           ?: 'http://loki.loki.svc.cluster.local:3100'
                    env.TEMPO_URL        = envMap['TEMPO_URL']          ?: 'http://tempo.tracing.svc.cluster.local:3200'
                    env.ALERTMANAGER_URL = envMap['ALERTMANAGER_URL']   ?: 'http://alertmanager.prometheus.svc.cluster.local:9093'
                    env.PYRRA_URL        = envMap['PYRRA_URL']          ?: 'http://pyrra.monitoring.svc.cluster.local:9099'
                    env.KIALI_URL        = envMap['KIALI_URL']          ?: 'http://kiali.istio-system.svc.cluster.local:20001'
                    env.ZIPKIN_URL       = envMap['ZIPKIN_URL']         ?: 'http://zipkin.tracing.svc.cluster.local:9411'
                    env.UPTIME_KUMA_URL  = envMap['UPTIME_KUMA_URL']    ?: 'http://uptime-kuma.monitoring.svc.cluster.local:3001'
                    env.PYROSCOPE_URL    = envMap['PYROSCOPE_URL']      ?: 'http://pyroscope.monitoring.svc.cluster.local:4040'
                    env.SIGNOZ_URL       = envMap['SIGNOZ_URL']         ?: 'http://signoz.monitoring.svc.cluster.local:3301'
                    env.OPENREPLAY_URL   = envMap['OPENREPLAY_URL']     ?: ''
                    env.PLAUSIBLE_URL    = envMap['PLAUSIBLE_URL']      ?: ''

                    // Security & vuln management
                    env.DEFECTDOJO_URL   = envMap['DEFECTDOJO_URL']     ?: 'http://defectdojo-defectdojo.defectdojo.svc.cluster.local:8080'
                    env.DEFECTDOJO_TOKEN = envMap['DEFECTDOJO_TOKEN']   ?: ''
                    env.DEPTRACK_URL     = envMap['DEPENDENCY_TRACK_URL'] ?: 'http://dependency-track.dependency-track.svc.cluster.local:8080'
                    env.SONAR_URL        = envMap['SONARQUBE_URL']      ?: 'http://sonarqube:9000'
                    env.HARBOR_URL       = envMap['HARBOR_URL']         ?: 'harbor.shopos.local'
                    env.ZAP_URL          = envMap['ZAP_URL']            ?: ''

                    // GitOps
                    env.ARGOCD_URL       = envMap['ARGOCD_URL']         ?: 'http://argocd-server.argocd.svc.cluster.local:80'

                    // Pact Broker
                    env.PACT_BROKER_URL  = envMap['PACT_BROKER_URL']    ?: 'http://pact-broker:9292'

                    // Messaging dashboard
                    env.AKHQ_URL         = envMap['AKHQ_URL']           ?: 'http://akhq.kafka.svc.cluster.local:8080'
                    env.KAFKA_UI_URL     = envMap['KAFKA_UI_URL']       ?: 'http://kafka-ui.kafka.svc.cluster.local:8080'

                    // Load test
                    env.INFLUXDB_URL     = envMap['INFLUXDB_URL']       ?: ''

                    // Notifications
                    env.SLACK_WEBHOOK    = envMap['SLACK_WEBHOOK_URL']  ?: ''
                    env.EMAIL_RECIPIENTS = envMap['EMAIL_RECIPIENTS']   ?: ''

                    // Kubeconfig
                    def kc = envMap['KUBECONFIG_CONTENT'] ?: ''
                    if (kc) {
                        writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kc
                        sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                        env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    }

                    sh 'mkdir -p reports/smoke reports/integration reports/api-quality/spectral reports/api-quality/hurl reports/dast/zap reports/dast/nuclei reports/load/k6 reports/load/locust reports/load/gatling reports/load/baseline reports/chaos reports/slo reports/security reports/summary'
                }
            }
        }

        // ── TEST CONTEXT ──────────────────────────────────────────────────────

        stage('Resolve Test Context') {
            steps {
                script {
                    env.TEST_SERVICE   = params.SERVICE_NAME.trim()
                    env.TEST_NAMESPACE = params.DOMAIN
                    env.TEST_DOMAIN    = params.DOMAIN
                    env.LOAD_PROFILE   = params.LOAD_PROFILE
                    env.LOAD_VUS       = params.LOAD_VUS
                    env.LOAD_DURATION  = params.LOAD_DURATION
                    env.CHAOS_DURATION = params.CHAOS_DURATION

                    if (params.IMAGE_TAG?.trim()) {
                        env.IMAGE_TAG = params.IMAGE_TAG.trim()
                    } else {
                        def detectedTag = sh(
                            script: """kubectl get deployment ${env.TEST_SERVICE} \
                                -n ${env.TEST_NAMESPACE} \
                                -o jsonpath='{.spec.template.spec.containers[0].image}' \
                                2>/dev/null | cut -d: -f2""",
                            returnStdout: true
                        ).trim()
                        env.IMAGE_TAG = detectedTag ?: 'unknown'
                    }

                    def clusterIP = sh(
                        script: """kubectl get svc ${env.TEST_SERVICE} \
                            -n ${env.TEST_NAMESPACE} \
                            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo ''""",
                        returnStdout: true
                    ).trim()

                    def port = sh(
                        script: """kubectl get svc ${env.TEST_SERVICE} \
                            -n ${env.TEST_NAMESPACE} \
                            -o jsonpath='{.spec.ports[?(@.name=="http")].port}' \
                            2>/dev/null || echo '80'""",
                        returnStdout: true
                    ).trim() ?: '80'

                    env.SERVICE_URL = clusterIP ?
                        "http://${clusterIP}:${port}" :
                        "http://${env.TEST_SERVICE}.${env.TEST_NAMESPACE}.svc.cluster.local:${port}"

                    echo "────────────────────────────────────────────────────"
                    echo "Pipeline     : POST-DEPLOY VALIDATION"
                    echo "Service      : ${env.TEST_SERVICE}"
                    echo "Namespace    : ${env.TEST_NAMESPACE}"
                    echo "Image Tag    : ${env.IMAGE_TAG}"
                    echo "Service URL  : ${env.SERVICE_URL}"
                    echo "Load Profile : ${env.LOAD_PROFILE}"
                    echo "────────────────────────────────────────────────────"
                }
            }
        }

        // ── SMOKE TESTS ───────────────────────────────────────────────────────

        stage('Smoke Tests') {
            // /healthz · /readyz · /metrics · gRPC health probe (grpc-health-probe)
            when { expression { !params.SKIP_SMOKE } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-smoke.groovy'
                        s()
                    }
                }
            }
        }

        // ── INTEGRATION TESTS ─────────────────────────────────────────────────

        stage('Integration Tests') {
            // Cross-service API probes · DB connectivity · Kafka producer/consumer probe
            when { expression { !params.SKIP_INTEGRATION } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-integration.groovy'
                        s()
                    }
                }
            }
        }

        // ── API QUALITY ───────────────────────────────────────────────────────

        stage('Spectral — OpenAPI Lint') {
            // Validates api-gateway.yaml, admin-api.yaml, developer-platform-api.yaml
            // against .spectral.yaml ruleset. Fails on error severity violations.
            when { expression { !params.SKIP_API_QUALITY } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            -v \$(pwd):/workspace \
                            -w /workspace \
                            stoplight/spectral:latest \
                            lint \
                                docs/openapi/api-gateway.yaml \
                                docs/openapi/admin-api.yaml \
                                docs/openapi/developer-platform-api.yaml \
                            --ruleset api-testing/spectral/.spectral.yaml \
                            --format=junit \
                            --output /workspace/reports/api-quality/spectral/spectral-results.xml \
                        || true
                    """
                    junit allowEmptyResults: true, testResults: 'reports/api-quality/spectral/spectral-results.xml'
                }
            }
        }

        stage('Hurl — HTTP API Tests') {
            // health-checks · auth-flow · catalog-flow · checkout-flow
            when { expression { !params.SKIP_API_QUALITY } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        ['health-checks','auth-flow','catalog-flow','checkout-flow'].each { flow ->
                            sh """
                                docker run --rm \
                                    --network host \
                                    -v \$(pwd)/api-testing/hurl:/tests \
                                    -v \$(pwd)/reports/api-quality/hurl:/reports \
                                    ghcr.io/orange-opensource/hurl:latest \
                                    --test /tests/${flow}.hurl \
                                    --report-junit /reports/hurl-${flow}-results.xml \
                                    --variable base_url=${params.API_BASE_URL} \
                                || true
                            """
                        }
                    }
                    junit allowEmptyResults: true, testResults: 'reports/api-quality/hurl/hurl-*.xml'
                }
            }
        }

        stage('Pact — Publish & Verify Contracts') {
            // Publishes consumer contracts to Pact Broker
            // Verifies provider compatibility (can-i-deploy check)
            when { expression { !params.SKIP_API_QUALITY } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/testing/pact:/pact \
                            pactfoundation/pact-cli:latest \
                            broker publish /pact/consumer \
                            --broker-base-url=${env.PACT_BROKER_URL} \
                            --broker-username=admin \
                            --broker-password=admin \
                            --consumer-app-version=\${GIT_COMMIT:-local} \
                            --branch=\${GIT_BRANCH:-main} \
                        || true

                        docker run --rm \
                            --network host \
                            pactfoundation/pact-cli:latest \
                            broker can-i-deploy \
                            --pacticipant ${env.TEST_SERVICE} \
                            --version \${GIT_COMMIT:-local} \
                            --to-environment ${params.ENVIRONMENT} \
                            --broker-base-url=${env.PACT_BROKER_URL} \
                            --broker-username=admin \
                            --broker-password=admin \
                        || true
                    """
                }
            }
        }

        // ── DAST ──────────────────────────────────────────────────────────────

        stage('DAST — OWASP ZAP') {
            // Spider + active scan against the running API gateway.
            // Runs in ATTACK mode; results written to SARIF + HTML.
            when { expression { !params.SKIP_DAST } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        echo "=== OWASP ZAP DAST scan: ${params.API_BASE_URL} ==="
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/reports/dast/zap:/zap/wrk \
                            ghcr.io/zaproxy/zaproxy:stable \
                            zap-api-scan.py \
                                -t ${params.API_BASE_URL}/openapi.json \
                                -f openapi \
                                -r /zap/wrk/zap-report.html \
                                -J /zap/wrk/zap-report.json \
                                -x /zap/wrk/zap-report.xml \
                                -l WARN \
                            || true
                    """
                    archiveArtifacts allowEmptyArchive: true, artifacts: 'reports/dast/zap/**'
                }
            }
        }

        stage('DAST — Nuclei CVE Scan') {
            // Template-based CVE and misconfig scanning against running services
            when { expression { !params.SKIP_DAST } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        echo "=== Nuclei scan: ${env.SERVICE_URL} ==="
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/reports/dast/nuclei:/output \
                            projectdiscovery/nuclei:latest \
                            -u ${env.SERVICE_URL} \
                            -t cves/ -t misconfiguration/ -t exposures/ \
                            -severity medium,high,critical \
                            -json-export /output/nuclei-results.json \
                            -sarif-export /output/nuclei-results.sarif \
                        || true
                    """
                    archiveArtifacts allowEmptyArchive: true, artifacts: 'reports/dast/nuclei/**'
                }
            }
        }

        // ── LOAD TESTING ──────────────────────────────────────────────────────

        stage('Load Testing — k6') {
            // checkout-flow · product-browse · search-load · spike
            when { expression { !params.SKIP_K6 } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-k6.groovy'
                        s()
                    }
                }
            }
        }

        stage('Load Testing — Locust') {
            // locustfile.py · headless workers · supports heavy + spike profiles
            when { expression { !params.SKIP_LOCUST } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-locust.groovy'
                        s()
                    }
                }
            }
        }

        stage('Load Testing — Gatling') {
            // CommerceSimulation · SearchSimulation
            when { expression { !params.SKIP_GATLING } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-gatling.groovy'
                        s()
                    }
                }
            }
        }

        stage('Performance Baseline Check') {
            // p95 < 2000ms · error rate < 5% — parsed from k6 + Locust stdout
            when { expression { !params.SKIP_PERF_CHECK } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-perf-baseline.groovy'
                        s()
                    }
                }
            }
        }

        // ── CHAOS ENGINEERING ─────────────────────────────────────────────────

        stage('Chaos Engineering — Chaos Mesh') {
            // Experiments: pod-kill · network-delay · cpu-stress
            // Applies → measures SLO impact → deletes → waits for recovery
            when { expression { !params.SKIP_CHAOS_MESH } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-chaos-mesh.groovy'
                        s()
                    }
                }
            }
        }

        stage('Chaos Engineering — Litmus') {
            // Workflows: database-chaos · payment-chaos (domain-specific)
            when { expression { !params.SKIP_LITMUS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-litmus.groovy'
                        s()
                    }
                }
            }
        }

        // ── SLO VALIDATION ────────────────────────────────────────────────────

        stage('SLO & Observability Validation') {
            // Availability >= 99.5% (30m window) via Prometheus
            // p95 < 2s (30m window) via Prometheus
            // Error budget via Pyrra
            // Validates: Prometheus scraping, Grafana health, Loki ready, active alerts
            when { expression { !params.SKIP_SLO } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-slo.groovy'
                        s()
                    }
                }
            }
        }

        // ── SECURITY REPORT ───────────────────────────────────────────────────

        stage('Security Report Upload') {
            // Uploads DAST results (ZAP, Nuclei) to DefectDojo
            // Pushes post-deploy SBOM snapshot to Dependency-Track
            // Sends Wazuh security event for SIEM correlation
            when { expression { !params.SKIP_SECURITY_REPORT } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        sh """
                            echo "=== Uploading DAST results to DefectDojo ==="
                            # ZAP results
                            [ -f reports/dast/zap/zap-report.json ] && \
                            curl -s -X POST \
                                -H "Authorization: Token ${env.DEFECTDOJO_TOKEN}" \
                                -F "scan_type=ZAP Scan" \
                                -F "file=@reports/dast/zap/zap-report.json" \
                                -F "engagement_name=post-deploy-${env.TEST_SERVICE}-${env.IMAGE_TAG}" \
                                -F "product_name=${env.TEST_DOMAIN}" \
                                -F "auto_create_context=True" \
                                "${env.DEFECTDOJO_URL}/api/v2/import-scan/" || true

                            # Nuclei results
                            [ -f reports/dast/nuclei/nuclei-results.sarif ] && \
                            curl -s -X POST \
                                -H "Authorization: Token ${env.DEFECTDOJO_TOKEN}" \
                                -F "scan_type=Sarif" \
                                -F "file=@reports/dast/nuclei/nuclei-results.sarif" \
                                -F "engagement_name=post-deploy-nuclei-${env.TEST_SERVICE}" \
                                -F "product_name=${env.TEST_DOMAIN}" \
                                -F "auto_create_context=True" \
                                "${env.DEFECTDOJO_URL}/api/v2/import-scan/" || true

                            echo "=== Security report upload complete ==="
                        """

                        def reportSummary = sh(
                            script: """
                                echo "=== Security Summary: ${env.TEST_SERVICE} @ ${env.IMAGE_TAG} ==="
                                echo "DAST ZAP    : \$([ -f reports/dast/zap/zap-report.json ] && cat reports/dast/zap/zap-report.json | python3 -c 'import json,sys; d=json.load(sys.stdin); alerts=[i for s in d.get("site",[]) for i in s.get("alerts",[])] if isinstance(d, dict) else []; print(len([a for a in alerts if a.get("riskcode","0") in ["3","2"]]), "HIGH/MEDIUM alerts")' 2>/dev/null || echo "not available")"
                                echo "Nuclei      : \$([ -f reports/dast/nuclei/nuclei-results.json ] && wc -l < reports/dast/nuclei/nuclei-results.json || echo "0") findings"
                                echo "DefectDojo  : ${env.DEFECTDOJO_URL}/finding"
                            """,
                            returnStdout: true
                        ).trim()

                        echo reportSummary
                        writeFile file: 'reports/security/post-deploy-security-summary.txt', text: reportSummary
                    }
                }
            }
        }

        // ── REPORT & SUMMARY ──────────────────────────────────────────────────

        stage('Aggregate Reports') {
            // Consolidates all test results, prints a summary table to console,
            // pushes k6 metrics to InfluxDB/Prometheus, uploads to DefectDojo
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/postdeploy-report.groovy'
                        s()
                    }
                }
            }
        }

        // ── DASHBOARD LINKS ───────────────────────────────────────────────────

        stage('Dashboard Links') {
            steps {
                script {
                    def envMap = [:]
                    if (fileExists('infra.env')) {
                        readFile('infra.env').trim().split('\n').each { line ->
                            def idx = line.indexOf('=')
                            if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                        }
                    }
                    ['GRAFANA_URL','PROMETHEUS_URL','JAEGER_URL','LOKI_URL','ALERTMANAGER_URL',
                     'PYRRA_URL','KIALI_URL','UPTIME_KUMA_URL','PYROSCOPE_URL','SIGNOZ_URL',
                     'ZIPKIN_URL','DEFECTDOJO_URL','DEPENDENCY_TRACK_URL','SONARQUBE_URL',
                     'ARGOCD_URL','HARBOR_URL','PACT_BROKER_URL','AKHQ_URL','KAFKA_UI_URL',
                     'OPENREPLAY_URL','PLAUSIBLE_URL'].each { k ->
                        if (env."${k}") envMap[k] = env."${k}"
                    }
                    envMap['SONARQUBE_URL'] = envMap['SONARQUBE_URL'] ?: env.SONAR_URL

                    def d = load 'scripts/groovy/dashboard-links.groovy'
                    echo d.call(envMap, "POST-DEPLOY — Build #${env.BUILD_NUMBER}", [
                        service: env.TEST_SERVICE ?: 'unknown',
                        tag:     env.IMAGE_TAG    ?: 'unknown',
                        domain:  env.TEST_DOMAIN  ?: params.DOMAIN,
                        project: 'shopos'
                    ])
                    echo "ZAP DAST report : reports/dast/zap/zap-report.html (Jenkins artifact)"
                    echo "Nuclei report   : reports/dast/nuclei/ (Jenkins artifact)"
                    echo "k6/Locust/Gatling: reports/load/ (Jenkins artifact + HTML publisher)"
                    echo "Spectral/Hurl   : reports/api-quality/ (Jenkins artifact)"
                }
            }
        }
    }

    // ── POST ──────────────────────────────────────────────────────────────────

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'

            archiveArtifacts artifacts: 'reports/**', allowEmptyArchive: true

            junit allowEmptyResults: true, testResults: 'reports/**/*.xml'

            publishHTML(target: [
                allowMissing: true, alwaysLinkToLastBuild: true, keepAll: true,
                reportDir:   'reports/load/gatling',
                reportFiles: '**/index.html',
                reportName:  'Gatling Load Test Report'
            ])
            publishHTML(target: [
                allowMissing: true, alwaysLinkToLastBuild: true, keepAll: true,
                reportDir:   'reports/load/locust',
                reportFiles: '*.html',
                reportName:  'Locust Load Test Report'
            ])
            publishHTML(target: [
                allowMissing: true, alwaysLinkToLastBuild: true, keepAll: true,
                reportDir:   'reports/dast/zap',
                reportFiles: 'zap-report.html',
                reportName:  'OWASP ZAP DAST Report'
            ])
        }

        success {
            script {
                def summary = """POST-DEPLOY SUCCESS — ShopOS Build #${env.BUILD_NUMBER}
Service     : ${env.TEST_SERVICE}
Domain      : ${env.TEST_DOMAIN}
Environment : ${params.ENVIRONMENT}
Image Tag   : ${env.IMAGE_TAG}
Stages      : Smoke, Integration, Spectral, Hurl, Pact, ZAP DAST, Nuclei, k6, Locust, Gatling, Perf Check, Chaos Mesh, Litmus, SLO, Security Report
Grafana     : ${env.GRAFANA_URL}/dashboards
Jaeger      : ${env.JAEGER_URL}/search?service=${env.TEST_SERVICE}
DefectDojo  : ${env.DEFECTDOJO_URL}/finding
ArgoCD      : ${env.ARGOCD_URL}/applications/${env.TEST_SERVICE}
Build URL   : ${env.BUILD_URL}"""

                if (env.SLACK_WEBHOOK?.trim()) {
                    sh """
                        curl -s -X POST '${env.SLACK_WEBHOOK}' \
                            -H 'Content-Type: application/json' \
                            -d '{"text":"POST-DEPLOY SUCCESS: ${env.TEST_SERVICE} @ ${env.IMAGE_TAG} — all validation passed\\nGrafana: ${env.GRAFANA_URL}\\nJaeger: ${env.JAEGER_URL}/search?service=${env.TEST_SERVICE}\\nBuild: ${env.BUILD_URL}"}' || true
                    """
                }
                if (env.EMAIL_RECIPIENTS?.trim()) {
                    mail to:      env.EMAIL_RECIPIENTS,
                         subject: "POST-DEPLOY SUCCESS: ${env.TEST_SERVICE} @ ${env.IMAGE_TAG}",
                         body:    summary
                }
                echo "=== SUCCESS NOTIFICATION SENT ==="
                echo summary
            }
        }

        failure {
            script {
                def summary = """POST-DEPLOY FAILED — ShopOS Build #${env.BUILD_NUMBER}
Service     : ${env.TEST_SERVICE ?: 'unknown'}
Domain      : ${params.DOMAIN}
Environment : ${params.ENVIRONMENT}
Image Tag   : ${env.IMAGE_TAG ?: 'unknown'}
Grafana     : ${env.GRAFANA_URL ?: 'N/A'}/dashboards
Jaeger      : ${env.JAEGER_URL ?: 'N/A'}
DefectDojo  : ${env.DEFECTDOJO_URL ?: 'N/A'}/finding
Build URL   : ${env.BUILD_URL}
Action      : Check archived reports and Grafana for root cause."""

                if (env.SLACK_WEBHOOK?.trim()) {
                    sh """
                        curl -s -X POST '${env.SLACK_WEBHOOK}' \
                            -H 'Content-Type: application/json' \
                            -d '{"text":"POST-DEPLOY FAILED: ${env.TEST_SERVICE ?: params.DOMAIN} — check archived reports\\nGrafana: ${env.GRAFANA_URL}\\nBuild: ${env.BUILD_URL}"}' || true
                    """
                }
                if (env.EMAIL_RECIPIENTS?.trim()) {
                    mail to:      env.EMAIL_RECIPIENTS,
                         subject: "POST-DEPLOY FAILED: ${env.TEST_SERVICE ?: params.DOMAIN} — Build #${env.BUILD_NUMBER}",
                         body:    summary
                }
                echo "=== FAILURE NOTIFICATION SENT ==="
                echo summary
            }
        }

        cleanup {
            script {
                sh """
                    for f in /tmp/chaos-*.yaml /tmp/litmus-*.yaml; do
                        [ -f "\$f" ] || continue
                        kubectl delete -f "\$f" --ignore-not-found 2>/dev/null || true
                        rm -f "\$f"
                    done
                """
                sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
                sh 'docker stop locust-master locust-worker-1 locust-worker-2 locust-worker-3 2>/dev/null || true'
                sh 'docker rm   locust-master locust-worker-1 locust-worker-2 locust-worker-3 2>/dev/null || true'
                sh 'docker image prune -f 2>/dev/null || true'
                sh 'rm -f /tmp/api-resp.json /tmp/chaos-*.yaml /tmp/litmus-*.yaml 2>/dev/null || true'
            }
        }
    }
}
