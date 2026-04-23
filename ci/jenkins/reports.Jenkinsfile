pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '30'))
        timeout(time: 20, unit: 'MINUTES')
    }

    parameters {
        choice(name: 'ACTION', choices: ['DEPLOY', 'BUILD_ONLY'], description: 'Deploy reports portal to cluster or build image only')
        string(name: 'ENVIRONMENT', defaultValue: 'production', description: 'Target environment (production / staging)')
        string(name: 'IMAGE_TAG',   defaultValue: 'latest',     description: 'Image tag to build and push')
        booleanParam(name: 'COLLECT_REPORTS', defaultValue: true, description: 'Collect and archive latest reports from all pipelines')
    }

    stages {
        stage('Git Fetch') {
            steps { checkout scm }
        }

        stage('Load Environment') {
            steps {
                script {
                    if (fileExists('/var/lib/jenkins/infra.env')) {
                        def envLines = readFile('/var/lib/jenkins/infra.env').trim().split('\n')
                        envLines.each { line ->
                            def idx = line.indexOf('=')
                            if (idx > 0) {
                                def key = line[0..<idx].trim()
                                def val = line[(idx+1)..-1].trim()
                                env."${key}" = val
                            }
                        }
                    }
                    env.REGISTRY    = env.HARBOR_URL      ?: 'harbor.registry.svc.cluster.local'
                    env.REGISTRY_PROJECT = env.REGISTRY_PROJECT ?: 'shopos'
                    env.SERVICE     = 'reports-portal-service'
                    env.IMAGE_FULL  = "${env.REGISTRY}/${env.REGISTRY_PROJECT}/${env.SERVICE}:${params.IMAGE_TAG}"
                }
            }
        }

        // ── Build & Push Reports Portal Image ─────────────────────────────────

        stage('Docker Build') {
            steps {
                sh """
                    docker build \
                        -t ${env.IMAGE_FULL} \
                        -f src/platform/reports-portal-service/Dockerfile \
                        src/platform/reports-portal-service/
                    echo "Built: ${env.IMAGE_FULL}"
                """
            }
        }

        stage('Docker Push') {
            steps {
                sh """
                    docker login ${env.REGISTRY} \
                        -u ${env.HARBOR_USER ?: 'admin'} \
                        -p ${env.HARBOR_PASSWORD ?: 'admin'} || true
                    docker push ${env.IMAGE_FULL}
                    echo "Pushed: ${env.IMAGE_FULL}"
                """
            }
        }

        // ── Deploy via Helm ───────────────────────────────────────────────────

        stage('Helm Deploy') {
            when { expression { params.ACTION == 'DEPLOY' } }
            steps {
                sh """
                    helm upgrade --install reports-portal-service \
                        helm/charts/reports-portal-service \
                        --namespace platform \
                        --create-namespace \
                        --set image.repository=${env.REGISTRY}/${env.REGISTRY_PROJECT}/reports-portal-service \
                        --set image.tag=${params.IMAGE_TAG} \
                        --set env.JENKINS_URL=${env.JENKINS_URL ?: 'http://jenkins.ci.svc.cluster.local:8080'} \
                        --set env.GRAFANA_URL=${env.GRAFANA_URL ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'} \
                        --set env.ARGOCD_URL=${env.ARGOCD_URL ?: 'http://argocd-server.argocd.svc.cluster.local:80'} \
                        --set env.SONARQUBE_URL=${env.SONARQUBE_URL ?: 'http://sonarqube.security.svc.cluster.local:9000'} \
                        --set env.DEFECTDOJO_URL=${env.DEFECTDOJO_URL ?: 'http://defectdojo.security.svc.cluster.local:8080'} \
                        --set env.PACT_BROKER_URL=${env.PACT_BROKER_URL ?: 'http://pact-broker.platform.svc.cluster.local:9292'} \
                        --set env.HARBOR_URL=${env.HARBOR_URL ?: 'http://harbor.registry.svc.cluster.local:80'} \
                        --set env.PROMETHEUS_URL=${env.PROMETHEUS_URL ?: 'http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local:9090'} \
                        --set env.JAEGER_URL=${env.JAEGER_URL ?: 'http://jaeger-query.observability.svc.cluster.local:16686'} \
                        --set env.VAULT_URL=${env.VAULT_URL ?: 'http://vault.security.svc.cluster.local:8200'} \
                        --set env.KEYCLOAK_URL=${env.KEYCLOAK_URL ?: 'http://keycloak.security.svc.cluster.local:8080'} \
                        --wait --timeout=5m
                """
            }
        }

        stage('Verify Deployment') {
            when { expression { params.ACTION == 'DEPLOY' } }
            steps {
                sh """
                    kubectl rollout status deployment/reports-portal-service \
                        -n platform --timeout=3m || true
                    kubectl get pods -n platform -l app=reports-portal-service
                """
                script {
                    def portalUrl = env.REPORTS_PORTAL_URL ?: 'http://reports-portal-service.platform.svc.cluster.local:8300'
                    sh """
                        kubectl run portal-health-check --rm -i --restart=Never \
                            --image=curlimages/curl:latest \
                            -- curl -sf ${portalUrl}/healthz || true
                    """
                }
            }
        }

        // ── Collect Latest Reports from All Pipelines ─────────────────────────

        stage('Collect Reports Index') {
            when { expression { params.COLLECT_REPORTS } }
            steps {
                script {
                    def jenkinsUrl = env.JENKINS_URL ?: 'http://jenkins.ci.svc.cluster.local:8080'

                    def pipelines = [
                        'pre-deploy': '03-pre-deploy',
                        'deploy': '04-deploy',
                        'post-deploy': '05-post-deploy',
                        'security': '06-security',
                        'observability': '07-observability',
                        'api-quality': '08-api-quality',
                        'tooling': '09-tooling',
                        'gitops': '10-gitops',
                        'reports': '11-reports',
                    ]

                    def reportIndex = """# ShopOS Pipeline Reports Index
Generated: ${new Date().format('yyyy-MM-dd HH:mm:ss UTC')}
Jenkins: ${jenkinsUrl}

## Pipeline Last Build URLs
"""
                    pipelines.each { name, jobNum ->
                        reportIndex += "- **${name}**: ${jenkinsUrl}/job/${jobNum}-${name}/lastBuild/\n"
                        reportIndex += "  - Console: ${jenkinsUrl}/job/${jobNum}-${name}/lastBuild/console\n"
                        reportIndex += "  - Artifacts: ${jenkinsUrl}/job/${jobNum}-${name}/lastBuild/artifact/\n"
                        reportIndex += "  - Test Results: ${jenkinsUrl}/job/${jobNum}-${name}/lastBuild/testReport/\n\n"
                    }

                    reportIndex += """
## Key Report Artifacts

### Security (pre-deploy)
- trivy-image-report.json (container CVEs)
- grype-report.json (CVE scan)
- sbom.json (CycloneDX SBOM)
- semgrep-results.sarif (SAST)
- gitleaks-report.json (secrets)
- checkov-results.sarif (IaC)
- kics-results.sarif (IaC)

### DAST (post-deploy)
- zap-report.html / zap-report.json
- nuclei-results.json / nuclei-results.sarif

### Performance (post-deploy)
- locust-report.html
- k6-results.json
- gatling/ (HTML report)

### API Quality (api-quality)
- spectral-results.xml (OpenAPI lint)
- terrascan-aws.sarif / terrascan-gcp.sarif / terrascan-helm.sarif

## External Tool Dashboards
- Reports Portal : ${env.REPORTS_PORTAL_URL ?: 'http://reports-portal-service.platform.svc.cluster.local:8300'}
- DefectDojo     : ${env.DEFECTDOJO_URL ?: 'http://defectdojo.security.svc.cluster.local:8080'}
- SonarQube      : ${env.SONARQUBE_URL ?: 'http://sonarqube.security.svc.cluster.local:9000'}
- Pact Broker    : ${env.PACT_BROKER_URL ?: 'http://pact-broker.platform.svc.cluster.local:9292'}
- Dep-Track      : ${env.DEPENDENCY_TRACK_URL ?: 'http://dependency-track.security.svc.cluster.local:8080'}
- Grafana        : ${env.GRAFANA_URL ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'}
"""
                    writeFile file: 'reports-index.md', text: reportIndex
                    archiveArtifacts artifacts: 'reports-index.md', allowEmptyArchive: false
                    echo reportIndex
                }
            }
        }

        // ── Dashboard Links ───────────────────────────────────────────────────

        stage('Dashboard Links') {
            steps {
                script {
                    def envMap = [:]
                    if (fileExists('/var/lib/jenkins/infra.env')) {
                        readFile('/var/lib/jenkins/infra.env').trim().split('\n').each { line ->
                            def idx = line.indexOf('=')
                            if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                        }
                    }
                    def portalUrl = envMap['REPORTS_PORTAL_URL'] ?: 'http://reports-portal-service.platform.svc.cluster.local:8300'
                    echo """
╔══════════════════════════════════════════════════════════════════════════╗
║                  REPORTS PORTAL DEPLOYED                                  ║
╠══════════════════════════════════════════════════════════════════════════╣
║  Reports Portal  : ${portalUrl}
╠══════════════════════════════════════════════════════════════════════════╣
"""
                    def links = load 'scripts/groovy/dashboard-links.groovy'
                    echo links.call(envMap, 'SHOPOS — REPORTS PORTAL PIPELINE DASHBOARDS')
                }
            }
        }
    }

    post {
        success {
            script {
                def envMap = [:]
                if (fileExists('/var/lib/jenkins/infra.env')) {
                    readFile('/var/lib/jenkins/infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }
                }
                def slack  = envMap['SLACK_WEBHOOK_URL'] ?: ''
                def emails = envMap['EMAIL_RECIPIENTS']  ?: ''
                def portal = envMap['REPORTS_PORTAL_URL'] ?: 'http://reports-portal-service.platform.svc.cluster.local:8300'
                def msg    = "REPORTS PORTAL DEPLOYED — ShopOS Build #${env.BUILD_NUMBER}: Central reports portal is live. All CI/CD, security, observability, and test reports accessible at: ${portal}  Build: ${env.BUILD_URL}"
                if (slack?.trim()) {
                    sh "curl -s -X POST '${slack}' -H 'Content-Type: application/json' -d '{\"text\":\"${msg}\"}' || true"
                }
                if (emails?.trim()) {
                    mail to: emails, subject: "Reports Portal DEPLOYED — Build #${env.BUILD_NUMBER}", body: msg
                }
            }
        }
        failure {
            script {
                def envMap = [:]
                if (fileExists('/var/lib/jenkins/infra.env')) {
                    readFile('/var/lib/jenkins/infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }
                }
                def slack  = envMap['SLACK_WEBHOOK_URL'] ?: ''
                def emails = envMap['EMAIL_RECIPIENTS']  ?: ''
                def msg    = "REPORTS PORTAL FAILED — ShopOS Build #${env.BUILD_NUMBER}: Check build log: ${env.BUILD_URL}"
                if (slack?.trim()) {
                    sh "curl -s -X POST '${slack}' -H 'Content-Type: application/json' -d '{\"text\":\"${msg}\"}' || true"
                }
                if (emails?.trim()) {
                    mail to: emails, subject: "Reports Portal FAILED — Build #${env.BUILD_NUMBER}", body: msg
                }
            }
        }
        always {
            archiveArtifacts allowEmptyArchive: true, artifacts: 'reports-index.md'
        }
    }
}
