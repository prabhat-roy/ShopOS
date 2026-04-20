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
            description: 'Service to test (e.g. order-service, checkout-service). Required.'
        )
        choice(
            name: 'DOMAIN',
            choices: ['commerce','platform','identity','catalog','supply-chain','financial',
                      'customer-experience','communications','content','analytics-ai','b2b',
                      'integrations','affiliate'],
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
            description: 'Image tag of the deployed build (used in reports). Auto-detected from K8s if blank.'
        )

        // ── Load testing ──────────────────────────────────────────────────────
        choice(
            name: 'LOAD_PROFILE',
            choices: ['medium','light','heavy','spike'],
            description: 'Load profile — light (10VU/2m) · medium (50VU/5m) · heavy (200VU/10m) · spike (500VU/30s)'
        )
        string(
            name: 'LOAD_VUS',
            defaultValue: '',
            description: 'Override virtual users (blank = use profile default)'
        )
        string(
            name: 'LOAD_DURATION',
            defaultValue: '',
            description: 'Override test duration e.g. 3m, 300s (blank = use profile default)'
        )

        // ── Chaos ─────────────────────────────────────────────────────────────
        string(
            name: 'CHAOS_DURATION',
            defaultValue: '2m',
            description: 'How long each chaos experiment runs before measuring recovery'
        )

        // ── Stage skip flags ──────────────────────────────────────────────────
        booleanParam(name: 'SKIP_SMOKE',        defaultValue: false, description: 'Skip smoke tests (health endpoints, gRPC probe)')
        booleanParam(name: 'SKIP_INTEGRATION',  defaultValue: false, description: 'Skip integration tests (cross-service API probes, DB + Kafka connectivity)')
        booleanParam(name: 'SKIP_K6',           defaultValue: false, description: 'Skip k6 load tests (checkout-flow, product-browse, search, spike)')
        booleanParam(name: 'SKIP_LOCUST',       defaultValue: false, description: 'Skip Locust load tests (single + distributed workers for heavy/spike profiles)')
        booleanParam(name: 'SKIP_GATLING',      defaultValue: false, description: 'Skip Gatling simulations (CommerceSimulation / SearchSimulation)')
        booleanParam(name: 'SKIP_PERF_CHECK',   defaultValue: false, description: 'Skip performance baseline check (p95 < 2s, error rate < 5%)')
        booleanParam(name: 'SKIP_CHAOS_MESH',   defaultValue: false, description: 'Skip Chaos Mesh experiments (pod-kill, network-delay, cpu-stress)')
        booleanParam(name: 'SKIP_LITMUS',       defaultValue: false, description: 'Skip Litmus chaos workflows (database-chaos, payment-chaos)')
        booleanParam(name: 'SKIP_SLO',          defaultValue: false, description: 'Skip SLO & observability validation (availability, p95, Prometheus, Grafana, alerts)')
    }

    stages {

        // ──────────────────────────────────────────────────────────────��───────
        stage('Git Fetch') {
            steps {
                checkout scm
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Load Environment') {
            steps {
                script {
                    if (!params.SERVICE_NAME?.trim()) {
                        error 'SERVICE_NAME is required'
                    }

                    // Kubeconfig
                    def kubeconfigContent = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }
                        ?.split('=', 2)?.last()
                    if (kubeconfigContent) {
                        sh "echo '${kubeconfigContent}' | base64 -d > ${env.WORKSPACE}/kubeconfig"
                        env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    }

                    // Parse infra.env
                    def envMap = [:]
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }

                    // Observability tool URLs
                    env.PROM_URL       = envMap['PROMETHEUS_URL']    ?: 'http://prometheus-prometheus.prometheus.svc.cluster.local:9090'
                    env.GRAFANA_URL    = envMap['GRAFANA_URL']        ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'
                    env.LOKI_URL       = envMap['LOKI_URL']           ?: 'http://loki-loki.loki.svc.cluster.local:3100'
                    env.PYRRA_URL      = envMap['PYRRA_URL']          ?: ''
                    env.INFLUXDB_URL   = envMap['INFLUXDB_URL']       ?: ''

                    // Reporting
                    env.DEFECTDOJO_URL   = envMap['DEFECTDOJO_URL']   ?: ''
                    env.DEFECTDOJO_TOKEN = envMap['DEFECTDOJO_TOKEN']  ?: ''

                    sh 'mkdir -p reports/smoke reports/integration reports/load/k6 reports/load/locust reports/load/gatling reports/load/baseline reports/chaos reports/slo reports/summary'
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Resolve Test Context') {
            steps {
                script {
                    env.TEST_SERVICE   = params.SERVICE_NAME.trim()
                    // Each service lives in its own namespace (= service name)
                    env.TEST_NAMESPACE = params.SERVICE_NAME.trim()
                    env.TEST_DOMAIN    = params.DOMAIN
                    env.LOAD_PROFILE   = params.LOAD_PROFILE
                    env.LOAD_VUS       = params.LOAD_VUS
                    env.LOAD_DURATION  = params.LOAD_DURATION
                    env.CHAOS_DURATION = params.CHAOS_DURATION

                    // Detect deployed image tag from running pod if not supplied
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

                    // Resolve in-cluster service URL
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

                    env.SERVICE_URL = clusterIP ? "http://${clusterIP}:${port}" \
                        : "http://${env.TEST_SERVICE}.${env.TEST_NAMESPACE}.svc.cluster.local:${port}"

                    echo "────────────────────────────────────────────────────"
                    echo "Service     : ${env.TEST_SERVICE}"
                    echo "Namespace   : ${env.TEST_NAMESPACE}"
                    echo "Image tag   : ${env.IMAGE_TAG}"
                    echo "Service URL : ${env.SERVICE_URL}"
                    echo "Load profile: ${env.LOAD_PROFILE}"
                    echo "────────────────────────────────────────────────────"
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Smoke Tests') {
            // /healthz · /readyz · /metrics · gRPC health probe
            when { expression { !params.SKIP_SMOKE } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-smoke.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Integration Tests') {
            // Cross-service API probes · DB connectivity · Kafka probe
            when { expression { !params.SKIP_INTEGRATION } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-integration.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Load Testing — k6') {
            // Scripts: checkout-flow.js · product-browse.js · search-load.js · spike
            when { expression { !params.SKIP_K6 } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-k6.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Load Testing — Locust') {
            // locustfile.py · headless · distributed workers for heavy/spike
            when { expression { !params.SKIP_LOCUST } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-locust.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Load Testing — Gatling') {
            // CommerceSimulation (commerce) · SearchSimulation (catalog)
            when { expression { !params.SKIP_GATLING } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-gatling.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Performance Baseline Check') {
            // p95 < 2000ms · error rate < 5% — parsed from k6 + Locust output
            when { expression { !params.SKIP_PERF_CHECK } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-perf-baseline.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Chaos Engineering — Chaos Mesh') {
            // pod-kill · network-delay · cpu-stress
            // Applies experiment → measures impact → deletes → waits for recovery
            when { expression { !params.SKIP_CHAOS_MESH } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-chaos-mesh.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Chaos Engineering — Litmus') {
            // database-chaos · payment-chaos (domain-specific)
            // Applies LitmusChaos workflow → polls for completion → captures verdict
            when { expression { !params.SKIP_LITMUS } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-litmus.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('SLO & Observability Validation') {
            // Availability >= 99.5% (30m window)
            // p95 < 2s (30m window) via Prometheus
            // Error budget via Pyrra
            // Prometheus scraping · active alerts · Grafana health · Loki ready
            when { expression { !params.SKIP_SLO } }
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-slo.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Report & Notify') {
            // Aggregate all JSON reports into a single summary
            // Upload to DefectDojo · push k6 metrics to InfluxDB
            // Print human-readable table to console
            steps {
                script {
                    def s = load 'scripts/groovy/postdeploy-report.groovy'
                    s()
                }
            }
        }
    }

    // ──────────────────────────────────────────────────────────────────────────
    post {
        always {
            // Archive all reports: load test results, Gatling HTML, chaos, SLO
            archiveArtifacts artifacts: 'reports/**', allowEmptyArchive: true

            // Publish Gatling HTML report (if plugin installed)
            publishHTML(target: [
                allowMissing: true,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: 'reports/load/gatling',
                reportFiles: '**/index.html',
                reportName: 'Gatling Load Test Report'
            ])

            // Publish Locust HTML report
            publishHTML(target: [
                allowMissing: true,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: 'reports/load/locust',
                reportFiles: '*.html',
                reportName: 'Locust Load Test Report'
            ])

            echo "Build #${env.BUILD_NUMBER} complete — all reports archived."
        }

        success {
            echo "POST-DEPLOY SUCCESS — ${env.TEST_SERVICE} @ ${env.IMAGE_TAG} passed all checks"
        }

        failure {
            echo "POST-DEPLOY FAILED — ${env.TEST_SERVICE} — check archived reports for details"
        }

        cleanup {
            script {
                echo "=== Cleanup ==="

                // Remove any leftover chaos experiments that may still be applied
                sh """
                    for f in /tmp/chaos-*.yaml /tmp/litmus-*.yaml; do
                        [ -f "\$f" ] || continue
                        kubectl delete -f "\$f" --ignore-not-found 2>/dev/null || true
                        rm -f "\$f"
                    done
                """

                // Remove kubeconfig written to workspace
                sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"

                // Remove Locust worker containers if they're still running
                sh """
                    docker stop locust-master locust-worker-1 locust-worker-2 locust-worker-3 2>/dev/null || true
                    docker rm   locust-master locust-worker-1 locust-worker-2 locust-worker-3 2>/dev/null || true
                """

                // Remove dangling docker images left by load-test containers
                sh 'docker image prune -f 2>/dev/null || true'

                // Remove temp report files
                sh 'rm -f /tmp/api-resp.json /tmp/chaos-*.yaml /tmp/litmus-*.yaml 2>/dev/null || true'

                echo "Cleanup complete."
            }
        }
    }
}
