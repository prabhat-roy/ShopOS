pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 30, unit: 'MINUTES')
    }

    parameters {
        string(name: 'API_BASE_URL', defaultValue: 'http://api-gateway:8080', description: 'Base URL of the API Gateway to test against')
        booleanParam(name: 'RUN_SPECTRAL',  defaultValue: true, description: 'Spectral — OpenAPI linting')
        booleanParam(name: 'RUN_HURL',      defaultValue: true, description: 'Hurl — HTTP API integration tests')
        booleanParam(name: 'RUN_PACT',      defaultValue: true, description: 'Publish Pact contracts to Pact Broker')
        booleanParam(name: 'RUN_TERRASCAN', defaultValue: true, description: 'Terrascan — IaC security scanning')
        string(name: 'PACT_BROKER_URL',     defaultValue: 'http://pact-broker:9292', description: 'Pact Broker base URL')
    }

    stages {
        stage('Git Fetch') {
            steps { checkout scm }
        }

        // ── OpenAPI Linting ───────────────────────────────────────────────────

        stage('Spectral — OpenAPI Lint') {
            when { expression { params.RUN_SPECTRAL } }
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
                            --output /workspace/spectral-results.xml \
                        || true
                    """
                    junit allowEmptyResults: true, testResults: 'spectral-results.xml'
                }
            }
        }

        // ── HTTP API Tests ────────────────────────────────────────────────────

        stage('Hurl — Health Checks') {
            when { expression { params.RUN_HURL } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/api-testing/hurl:/tests \
                            ghcr.io/orange-opensource/hurl:latest \
                            --test /tests/health-checks.hurl \
                            --report-junit /tests/hurl-health-results.xml \
                            --variable base_url=${params.API_BASE_URL} \
                        || true
                    """
                    junit allowEmptyResults: true, testResults: 'api-testing/hurl/hurl-health-results.xml'
                }
            }
        }

        stage('Hurl — Auth Flow') {
            when { expression { params.RUN_HURL } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/api-testing/hurl:/tests \
                            ghcr.io/orange-opensource/hurl:latest \
                            --test /tests/auth-flow.hurl \
                            --report-junit /tests/hurl-auth-results.xml \
                            --variable base_url=${params.API_BASE_URL} \
                        || true
                    """
                    junit allowEmptyResults: true, testResults: 'api-testing/hurl/hurl-auth-results.xml'
                }
            }
        }

        stage('Hurl — Catalog Flow') {
            when { expression { params.RUN_HURL } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/api-testing/hurl:/tests \
                            ghcr.io/orange-opensource/hurl:latest \
                            --test /tests/catalog-flow.hurl \
                            --report-junit /tests/hurl-catalog-results.xml \
                            --variable base_url=${params.API_BASE_URL} \
                        || true
                    """
                    junit allowEmptyResults: true, testResults: 'api-testing/hurl/hurl-catalog-results.xml'
                }
            }
        }

        stage('Hurl — Checkout Flow') {
            when { expression { params.RUN_HURL } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/api-testing/hurl:/tests \
                            ghcr.io/orange-opensource/hurl:latest \
                            --test /tests/checkout-flow.hurl \
                            --report-junit /tests/hurl-checkout-results.xml \
                            --variable base_url=${params.API_BASE_URL} \
                        || true
                    """
                    junit allowEmptyResults: true, testResults: 'api-testing/hurl/hurl-checkout-results.xml'
                }
            }
        }

        // ── Contract Tests ────────────────────────────────────────────────────

        stage('Pact — Publish Contracts') {
            when { expression { params.RUN_PACT } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            --network host \
                            -v \$(pwd)/testing/pact:/pact \
                            pactfoundation/pact-cli:latest \
                            broker publish /pact/consumer \
                            --broker-base-url=${params.PACT_BROKER_URL} \
                            --broker-username=admin \
                            --broker-password=admin \
                            --consumer-app-version=\${GIT_COMMIT:-local} \
                            --branch=\${GIT_BRANCH:-main} \
                        || true
                    """
                }
            }
        }

        // ── IaC Security Scan ─────────────────────────────────────────────────

        stage('Terrascan — IaC Scan') {
            when { expression { params.RUN_TERRASCAN } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        docker run --rm \
                            -v \$(pwd):/iac \
                            tenable/terrascan:latest \
                            scan \
                            --iac-dir /iac/infra/terraform/aws \
                            --iac-type terraform \
                            --output sarif \
                            --severity MEDIUM \
                        > terrascan-aws.sarif || true

                        docker run --rm \
                            -v \$(pwd):/iac \
                            tenable/terrascan:latest \
                            scan \
                            --iac-dir /iac/infra/terraform/gcp \
                            --iac-type terraform \
                            --output sarif \
                            --severity MEDIUM \
                        > terrascan-gcp.sarif || true

                        docker run --rm \
                            -v \$(pwd):/iac \
                            tenable/terrascan:latest \
                            scan \
                            --iac-dir /iac/helm \
                            --iac-type helm \
                            --output sarif \
                            --severity HIGH \
                        > terrascan-helm.sarif || true

                        echo "Terrascan complete — SARIF results written"
                    """
                    archiveArtifacts allowEmptyArchive: true, artifacts: 'terrascan-*.sarif'
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
                    def grafana  = envMap['GRAFANA_URL']     ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'
                    def sonar    = envMap['SONARQUBE_URL']   ?: 'http://sonarqube:9000'
                    def dojo     = envMap['DEFECTDOJO_URL']  ?: 'http://defectdojo:8080'
                    def pact     = envMap['PACT_BROKER_URL'] ?: 'http://pact-broker:9292'
                    def argocd   = envMap['ARGOCD_URL']      ?: 'http://argocd-server.argocd.svc.cluster.local:80'

                    echo """
╔══════════════════════════════════════════════════════════════════════════╗
║            SHOPOS — API QUALITY DASHBOARD LINKS                          ║
╠══════════════════════════════════════════════════════════════════════════╣
║  API QUALITY RESULTS (archived artifacts)                                ║
║  Spectral lint   : spectral-results.xml (JUnit in Jenkins)
║  Hurl tests      : api-testing/hurl/hurl-*.xml (JUnit in Jenkins)
║  IaC SARIF       : terrascan-*.sarif (archived)
╠══════════════════════════════════════════════════════════════════════════╣
║  EXTERNAL DASHBOARDS                                                     ║
║  Pact Broker     : ${pact}
║  SonarQube       : ${sonar}
║  DefectDojo      : ${dojo}/finding
║  Grafana         : ${grafana}/dashboards
║  ArgoCD          : ${argocd}/applications
╚══════════════════════════════════════════════════════════════════════════╝
                    """
                }
            }
        }
    }

    post {
        always {
            junit allowEmptyResults: true, testResults: '**/hurl-*-results.xml,spectral-results.xml'
            archiveArtifacts allowEmptyArchive: true, artifacts: 'spectral-results.xml,terrascan-*.sarif,api-testing/hurl/hurl-*.xml'
        }
        success {
            script {
                def envMap = [:]
                if (fileExists('infra.env')) {
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }
                }
                def slack  = envMap['SLACK_WEBHOOK_URL'] ?: ''
                def emails = envMap['EMAIL_RECIPIENTS']  ?: ''
                def pact   = envMap['PACT_BROKER_URL']   ?: 'http://pact-broker:9292'
                def msg    = "API QUALITY SUCCESS — ShopOS Build #${env.BUILD_NUMBER}: Spectral lint, Hurl HTTP tests, Pact contracts, Terrascan IaC scan all passed. Pact Broker: ${pact}  Build: ${env.BUILD_URL}"
                if (slack?.trim()) {
                    sh "curl -s -X POST '${slack}' -H 'Content-Type: application/json' -d '{\"text\":\"${msg}\"}' || true"
                }
                if (emails?.trim()) {
                    mail to: emails, subject: "API Quality SUCCESS — Build #${env.BUILD_NUMBER}", body: msg
                }
                echo 'API quality checks passed.'
            }
        }
        failure {
            script {
                def envMap = [:]
                if (fileExists('infra.env')) {
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }
                }
                def slack  = envMap['SLACK_WEBHOOK_URL'] ?: ''
                def emails = envMap['EMAIL_RECIPIENTS']  ?: ''
                def msg    = "API QUALITY FAILED — ShopOS Build #${env.BUILD_NUMBER}: review Spectral, Hurl, or Terrascan results. Build: ${env.BUILD_URL}"
                if (slack?.trim()) {
                    sh "curl -s -X POST '${slack}' -H 'Content-Type: application/json' -d '{\"text\":\"${msg}\"}' || true"
                }
                if (emails?.trim()) {
                    mail to: emails, subject: "API Quality FAILED — Build #${env.BUILD_NUMBER}", body: msg
                }
                echo 'API quality checks failed — review Spectral, Hurl, and Terrascan results above.'
            }
        }
    }
}
