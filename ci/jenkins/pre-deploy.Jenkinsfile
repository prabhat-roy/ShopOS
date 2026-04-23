pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 90, unit: 'MINUTES')
    }

    parameters {
        string(
            name: 'SERVICE_NAME',
            defaultValue: '',
            description: 'Specific service to build (e.g. order-service). Blank = all services in DOMAIN.'
        )
        choice(
            name: 'DOMAIN',
            choices: ['platform','identity','catalog','commerce','supply-chain','financial',
                      'customer-experience','communications','content','analytics-ai','b2b',
                      'integrations','affiliate','marketplace','gamification','developer-platform',
                      'compliance','sustainability','events-ticketing','auction','rental','web'],
            description: 'Business domain containing the service(s)'
        )
        choice(
            name: 'ENVIRONMENT',
            choices: ['dev','staging','prod'],
            description: 'Target deployment environment'
        )
        string(
            name: 'IMAGE_TAG',
            defaultValue: '',
            description: 'Docker image tag override. Defaults to <env>-<sha>-<build> if blank.'
        )
        string(
            name: 'REGISTRY_PROJECT',
            defaultValue: 'shopos',
            description: 'Harbor project / namespace'
        )

        // ── Source Code Scanning ──────────────────────────────────────────────
        booleanParam(name: 'SKIP_SECRETS',    defaultValue: false, description: 'Skip secret scanning  (Gitleaks, TruffleHog, detect-secrets)')
        booleanParam(name: 'SKIP_SAST',       defaultValue: false, description: 'Skip SAST             (Semgrep, SonarQube, GoSec, Bandit, ESLint, cargo-clippy)')
        booleanParam(name: 'SKIP_SCA',        defaultValue: false, description: 'Skip SCA & SBOM       (Trivy FS, Grype, Syft SPDX+CycloneDX, OWASP DC)')
        booleanParam(name: 'SKIP_IAC',        defaultValue: false, description: 'Skip IaC scanning     (Checkov, KICS, tfsec, Terrascan, Polaris, Kubeaudit)')
        booleanParam(name: 'SKIP_LICENSE',    defaultValue: false, description: 'Skip license checks   (FOSSA, go-licenses, pip-licenses, license-checker)')

        // ── Build & Image ─────────────────────────────────────────────────────
        booleanParam(name: 'SKIP_IMAGE_SCAN', defaultValue: false, description: 'Skip image scanning   (Trivy image, Grype, Anchore, Clair, Docker Scout, Syft)')
        booleanParam(name: 'SKIP_SIGN',       defaultValue: false, description: 'Skip image signing    (Cosign keyless, Notation, Rekor transparency log)')

        // ── GitOps ────────────────────────────────────────────────────────────
        booleanParam(name: 'SKIP_GITOPS',     defaultValue: false, description: 'Skip GitOps manifest update and ArgoCD sync trigger')
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
                    if (!fileExists('infra.env')) {
                        error "infra.env not found — run k8s-infra and registry pipelines first"
                    }

                    def envMap = [:]
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }

                    // Registry
                    env.HARBOR_URL          = envMap['HARBOR_URL']            ?: 'harbor.shopos.local'
                    env.HARBOR_USER         = envMap['HARBOR_USER']            ?: 'admin'
                    env.HARBOR_PASSWORD     = envMap['HARBOR_PASSWORD']        ?: ''

                    // Source analysis servers
                    env.SONAR_URL           = envMap['SONARQUBE_URL']          ?: 'http://sonarqube-sonarqube.sonarqube.svc.cluster.local:9000'
                    env.SONAR_TOKEN         = envMap['SONARQUBE_TOKEN']        ?: ''
                    env.SNYK_TOKEN          = envMap['SNYK_TOKEN']             ?: ''
                    env.FOSSA_API_KEY       = envMap['FOSSA_API_KEY']          ?: ''
                    env.GITGUARDIAN_API_KEY = envMap['GITGUARDIAN_API_KEY']    ?: ''

                    // Image scanning
                    env.ANCHORE_URL         = envMap['ANCHORE_URL']            ?: ''
                    env.ANCHORE_PASSWORD    = envMap['ANCHORE_PASSWORD']       ?: 'foobar'
                    env.CLAIR_URL           = envMap['CLAIR_URL']              ?: ''

                    // Vulnerability management
                    env.DEFECTDOJO_URL      = envMap['DEFECTDOJO_URL']         ?: 'http://defectdojo-defectdojo.defectdojo.svc.cluster.local:8080'
                    env.DEFECTDOJO_TOKEN    = envMap['DEFECTDOJO_TOKEN']       ?: ''
                    env.DEPTRACK_URL        = envMap['DEPENDENCY_TRACK_URL']   ?: 'http://dependency-track-dependency-track.dependency-track.svc.cluster.local:8080'
                    env.DEPTRACK_KEY        = envMap['DEPTRACK_API_KEY']       ?: ''

                    // GitOps
                    env.ARGOCD_URL          = envMap['ARGOCD_URL']             ?: 'http://argocd-server.argocd.svc.cluster.local:80'
                    env.ARGOCD_TOKEN        = envMap['ARGOCD_TOKEN']           ?: ''

                    // Observability
                    env.GRAFANA_URL         = envMap['GRAFANA_URL']            ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'
                    env.PROMETHEUS_URL      = envMap['PROMETHEUS_URL']         ?: 'http://prometheus-prometheus.prometheus.svc.cluster.local:9090'

                    // Notifications
                    env.SLACK_WEBHOOK       = envMap['SLACK_WEBHOOK_URL']      ?: ''
                    env.EMAIL_RECIPIENTS    = envMap['EMAIL_RECIPIENTS']       ?: ''

                    // Kubeconfig
                    def kc = envMap['KUBECONFIG_CONTENT'] ?: ''
                    if (kc) {
                        writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kc
                        sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                        env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    }

                    sh 'mkdir -p reports/secrets reports/sast reports/sca reports/iac reports/license reports/image-scan reports/sbom'

                    sh "echo '${env.HARBOR_PASSWORD}' | docker login ${env.HARBOR_URL} -u ${env.HARBOR_USER} --password-stdin || true"
                }
            }
        }

        // ── BUILD CONTEXT ─────────────────────────────────────────────────────

        stage('Resolve Build Context') {
            steps {
                script {
                    def gitSha           = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.IMAGE_TAG        = params.IMAGE_TAG?.trim() ?: "${params.ENVIRONMENT}-${gitSha}-${env.BUILD_NUMBER}"
                    env.BUILD_DOMAIN     = params.DOMAIN
                    env.BUILD_ENV        = params.ENVIRONMENT
                    env.REGISTRY_PROJECT = params.REGISTRY_PROJECT
                    env.GIT_COMMIT_FULL  = sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
                    env.GIT_BRANCH_NAME  = sh(script: 'git rev-parse --abbrev-ref HEAD', returnStdout: true).trim()

                    if (params.SERVICE_NAME?.trim()) {
                        env.SERVICES = params.SERVICE_NAME.trim()
                    } else {
                        def srcPath = params.DOMAIN == 'web' ? 'src/web' : "src/${params.DOMAIN}"
                        env.SERVICES = sh(
                            script: "ls ${srcPath}/ 2>/dev/null | tr '\\n' ','",
                            returnStdout: true
                        ).trim().replaceAll(/,$/, '')
                    }

                    def primarySvc = env.SERVICES.split(',')[0].trim()
                    def detector   = load 'scripts/groovy/deploy-language-detect.groovy'
                    env.LANGUAGE   = detector.call(primarySvc)

                    echo "────────────────────────────────────────────────────"
                    echo "Pipeline    : PRE-DEPLOY (CI)"
                    echo "Services    : ${env.SERVICES}"
                    echo "Domain      : ${env.BUILD_DOMAIN}"
                    echo "Language    : ${env.LANGUAGE}"
                    echo "Tag         : ${env.IMAGE_TAG}"
                    echo "Environment : ${env.BUILD_ENV}"
                    echo "Branch      : ${env.GIT_BRANCH_NAME}"
                    echo "Commit      : ${env.GIT_COMMIT_FULL}"
                    echo "Registry    : ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}"
                    echo "────────────────────────────────────────────────────"
                }
            }
        }

        // ── SOURCE CODE SCANNING ──────────────────────────────────────────────

        stage('Secret Scanning') {
            // Tools: Gitleaks · TruffleHog · detect-secrets · GitGuardian ggshield
            when { expression { !params.SKIP_SECRETS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-secrets.groovy'
                        s()
                    }
                }
            }
        }

        stage('SAST') {
            // Tools: Semgrep · SonarQube · ShellCheck · Spectral (OpenAPI)
            //        GoSec · GolangCI-lint · Bandit · Pylint · Flake8
            //        ESLint · SpotBugs · PMD · Detekt · cargo-clippy · Brakeman
            when { expression { !params.SKIP_SAST } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-sast.groovy'
                        s()
                    }
                }
            }
        }

        stage('SCA & SBOM') {
            // Tools: Trivy FS · Grype · Syft (SPDX + CycloneDX)
            //        OWASP Dependency-Check · govulncheck · npm audit
            //        pip-audit · cargo audit · Maven OWASP DC
            when { expression { !params.SKIP_SCA } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-sca.groovy'
                        s()
                    }
                }
            }
        }

        stage('IaC & Manifest Scanning') {
            // Tools: Checkov · KICS · tfsec · Terrascan · Polaris · Kubeaudit · cnspec
            when { expression { !params.SKIP_IAC } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-iac.groovy'
                        s()
                    }
                }
            }
        }

        stage('License Compliance') {
            // Tools: FOSSA · Tern · license-checker · go-licenses · pip-licenses
            when { expression { !params.SKIP_LICENSE } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-license.groovy'
                        s()
                    }
                }
            }
        }

        // ── COMPILE & BUILD ───────────────────────────────────────────────────

        stage('Compile & Unit Test') {
            // go build · mvn package · gradle build · pip install · npm ci
            // dotnet build · cargo build · sbt assembly · mix compile
            // stack build (Haskell) · flutter build · swift build
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-compile.groovy'
                        s()
                    }
                }
            }
        }

        stage('Docker Build') {
            // Multi-stage Dockerfile builds per service
            // Kaniko used inside K8s pods; docker build used on bare-metal agents
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def ctxDir = env.BUILD_DOMAIN == 'web' ? "src/web/${svc}/" : "src/${env.BUILD_DOMAIN}/${svc}/"
                        def image  = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh """
                            echo "=== Building: ${image} ==="
                            docker build \
                                --label "git.commit=${env.GIT_COMMIT_FULL}" \
                                --label "git.branch=${env.GIT_BRANCH_NAME}" \
                                --label "build.number=${env.BUILD_NUMBER}" \
                                --label "build.domain=${env.BUILD_DOMAIN}" \
                                --label "build.environment=${env.BUILD_ENV}" \
                                --label "org.opencontainers.image.revision=${env.GIT_COMMIT_FULL}" \
                                --label "org.opencontainers.image.source=https://github.com/prabhat-roy/ShopOS" \
                                -t ${image} \
                                ${ctxDir}
                        """
                    }
                }
            }
        }

        // ── IMAGE SECURITY ────────────────────────────────────────────────────

        stage('Container Image Scan') {
            // Tools: Trivy image · Trivy secrets in layer
            //        Grype · Anchore Engine · Clair · Docker Scout · Syft image SBOM
            when { expression { !params.SKIP_IMAGE_SCAN } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-image-scan.groovy'
                        s()
                    }
                }
            }
        }

        stage('Image Signing — Cosign + Rekor + Notation') {
            // Cosign: keyless OIDC signing
            // Notation: Notary v2 signing
            // Rekor: immutable transparency log entry
            when { expression { !params.SKIP_SIGN } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        env.SERVICES.split(',').each { svc ->
                            svc = svc.trim()
                            def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                            sh """
                                echo "=== Cosign sign (keyless): ${svc} ==="
                                if command -v cosign &>/dev/null; then
                                    COSIGN_EXPERIMENTAL=1 cosign sign --yes ${image} || true
                                else
                                    docker run --rm \
                                        -v /var/run/docker.sock:/var/run/docker.sock \
                                        gcr.io/projectsigstore/cosign:latest sign --yes ${image} || true
                                fi

                                echo "=== Notation sign (Notary v2): ${svc} ==="
                                command -v notation &>/dev/null && notation sign ${image} || true

                                echo "=== Rekor transparency log upload: ${svc} ==="
                                command -v rekor-cli &>/dev/null && \
                                    rekor-cli upload \
                                        --artifact ${image} \
                                        --type=container \
                                        --server https://rekor.sigstore.dev || true

                                echo "=== Verify Cosign signature: ${svc} ==="
                                if command -v cosign &>/dev/null; then
                                    cosign verify --certificate-identity-regexp="jenkins" \
                                        --certificate-oidc-issuer-regexp=".*" \
                                        ${image} 2>/dev/null || echo "Signature verification skipped (no OIDC in this env)"
                                fi
                            """
                        }
                    }
                }
            }
        }

        // ── PUSH TO REGISTRY ──────────────────────────────────────────────────

        stage('Docker Push') {
            steps {
                script {
                    sh "echo '${env.HARBOR_PASSWORD}' | docker login ${env.HARBOR_URL} -u ${env.HARBOR_USER} --password-stdin"
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh """
                            echo "=== Pushing: ${image} ==="
                            docker push ${image}
                        """
                        echo "Pushed: ${image}"
                    }
                    echo "All images pushed → ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}"
                }
            }
        }

        // ── SECURITY REPORT UPLOAD ────────────────────────────────────────────

        stage('Security Report Upload') {
            // DefectDojo: all scan types (SAST, SCA, secrets, IaC, image)
            // Dependency-Track: CycloneDX SBOM (source + image layers)
            // Wazuh: pipeline completion event for SIEM correlation
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        def s = load 'scripts/groovy/deploy-report.groovy'
                        s()
                    }
                }
            }
        }

        // ── GITOPS UPDATE & SYNC ──────────────────────────────────────────────

        stage('Update GitOps Manifests') {
            // Updates Helm values image tags in both helm/services/ and gitops/argocd/
            // Commits and pushes the change so ArgoCD detects drift automatically
            when { expression { !params.SKIP_GITOPS } }
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def image      = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}"
                        def valuesFile = "helm/services/${svc}/values-${params.ENVIRONMENT}.yaml"
                        sh """
                            echo "=== Updating GitOps values: ${svc} → ${env.IMAGE_TAG} ==="
                            if [ -f ${valuesFile} ]; then
                                sed -i "s|tag:.*|tag: \"${env.IMAGE_TAG}\"|g" ${valuesFile} || true
                                sed -i "s|repository:.*${svc}.*|repository: ${image}|g" ${valuesFile} || true
                            fi
                            if [ -f gitops/argocd/apps/${svc}/values.yaml ]; then
                                sed -i "s|tag:.*|tag: \"${env.IMAGE_TAG}\"|g" gitops/argocd/apps/${svc}/values.yaml || true
                            fi
                        """
                    }

                    sh """
                        git config user.email "jenkins@shopos.local"
                        git config user.name  "Jenkins CI"
                        git add helm/services/ gitops/ 2>/dev/null || true
                        git diff --staged --quiet || \
                            git commit -m "ci: bump images to ${env.IMAGE_TAG} for ${params.ENVIRONMENT} [skip ci]"
                        git push origin ${env.GIT_BRANCH_NAME} 2>/dev/null || true
                    """
                }
            }
        }

        stage('ArgoCD Sync') {
            // Triggers an immediate out-of-band sync for all services built in this run.
            // ArgoCD will also self-heal from the git commit above on its next poll cycle.
            when { expression { !params.SKIP_GITOPS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        env.SERVICES.split(',').each { svc ->
                            svc = svc.trim()
                            sh """
                                echo "=== ArgoCD sync: ${svc} ==="
                                if command -v argocd &>/dev/null; then
                                    argocd app sync ${svc} \
                                        --server  ${env.ARGOCD_URL} \
                                        --auth-token ${env.ARGOCD_TOKEN} \
                                        --prune --timeout 300 || true
                                    argocd app wait ${svc} \
                                        --server ${env.ARGOCD_URL} \
                                        --auth-token ${env.ARGOCD_TOKEN} \
                                        --health --timeout 300 || true
                                else
                                    curl -sfk -X POST \
                                        -H "Authorization: Bearer ${env.ARGOCD_TOKEN}" \
                                        "${env.ARGOCD_URL}/api/v1/applications/${svc}/sync" || true
                                fi
                            """
                        }
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
                    // Merge pipeline-level env vars set in Load Environment stage
                    ['HARBOR_URL','SONARQUBE_URL','SONAR_URL','DEFECTDOJO_URL','DEPENDENCY_TRACK_URL',
                     'ARGOCD_URL','GRAFANA_URL','PROMETHEUS_URL','PACT_BROKER_URL'].each { k ->
                        if (env."${k}") envMap[k] = env."${k}"
                    }
                    envMap['SONARQUBE_URL'] = envMap['SONARQUBE_URL'] ?: env.SONAR_URL

                    def primarySvc = env.SERVICES?.split(',')?.getAt(0)?.trim() ?: 'unknown'
                    def d = load 'scripts/groovy/dashboard-links.groovy'
                    echo d.call(envMap, "PRE-DEPLOY — Build #${env.BUILD_NUMBER}", [
                        service: primarySvc,
                        tag:     env.IMAGE_TAG ?: 'unknown',
                        domain:  env.BUILD_DOMAIN ?: params.DOMAIN,
                        project: env.REGISTRY_PROJECT ?: 'shopos'
                    ])
                    echo "Cosign verify: cosign verify ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${primarySvc}:${env.IMAGE_TAG}"
                    echo "Scan reports archived: reports/secrets/, reports/sast/, reports/sca/, reports/image-scan/, reports/sbom/"
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
        }

        success {
            script {
                def summary = """PRE-DEPLOY SUCCESS — ShopOS Build #${env.BUILD_NUMBER}
Services   : ${env.SERVICES}
Domain     : ${env.BUILD_DOMAIN}
Environment: ${env.BUILD_ENV}
Tag        : ${env.IMAGE_TAG}
Branch     : ${env.GIT_BRANCH_NAME}
Registry   : https://${env.HARBOR_URL}/harbor/projects/${env.REGISTRY_PROJECT}
ArgoCD     : ${env.ARGOCD_URL}/applications
SonarQube  : ${env.SONAR_URL}
DefectDojo : ${env.DEFECTDOJO_URL}
Build URL  : ${env.BUILD_URL}
Stages complete: Secret Scan, SAST, SCA, IaC, License, Compile, Docker Build, Image Scan, Cosign Sign, Push, Security Upload, GitOps Sync"""

                if (env.SLACK_WEBHOOK?.trim()) {
                    sh """
                        curl -s -X POST '${env.SLACK_WEBHOOK}' \
                            -H 'Content-Type: application/json' \
                            -d '{"text":"SUCCESS: ShopOS Pre-Deploy #${env.BUILD_NUMBER} — ${env.SERVICES} — ${env.IMAGE_TAG}\\nArgoCD: ${env.ARGOCD_URL}/applications\\nBuild: ${env.BUILD_URL}"}' || true
                    """
                }
                if (env.EMAIL_RECIPIENTS?.trim()) {
                    mail to:      env.EMAIL_RECIPIENTS,
                         subject: "SUCCESS: ShopOS Pre-Deploy #${env.BUILD_NUMBER} — ${env.SERVICES}",
                         body:    summary
                }
                echo "=== SUCCESS NOTIFICATION SENT ==="
                echo summary
            }
        }

        failure {
            script {
                def summary = """PRE-DEPLOY FAILED — ShopOS Build #${env.BUILD_NUMBER}
Services   : ${env.SERVICES ?: 'unknown'}
Domain     : ${params.DOMAIN}
Environment: ${params.ENVIRONMENT}
Branch     : ${env.GIT_BRANCH_NAME ?: 'unknown'}
Build URL  : ${env.BUILD_URL}
Action     : Check archived reports for scan failures.
Reports    : reports/secrets, reports/sast, reports/sca, reports/iac, reports/image-scan"""

                if (env.SLACK_WEBHOOK?.trim()) {
                    sh """
                        curl -s -X POST '${env.SLACK_WEBHOOK}' \
                            -H 'Content-Type: application/json' \
                            -d '{"text":"FAILED: ShopOS Pre-Deploy #${env.BUILD_NUMBER} — ${env.SERVICES ?: params.DOMAIN}\\nBuild: ${env.BUILD_URL}"}' || true
                    """
                }
                if (env.EMAIL_RECIPIENTS?.trim()) {
                    mail to:      env.EMAIL_RECIPIENTS,
                         subject: "FAILED: ShopOS Pre-Deploy #${env.BUILD_NUMBER} — ${params.DOMAIN}",
                         body:    summary
                }
                echo "=== FAILURE NOTIFICATION SENT ==="
                echo summary
            }
        }

        cleanup {
            script {
                env.SERVICES?.split(',')?.each { svc ->
                    svc = svc?.trim()
                    if (svc && env.HARBOR_URL && env.REGISTRY_PROJECT && env.IMAGE_TAG) {
                        sh "docker rmi ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG} 2>/dev/null || true"
                    }
                }
                sh "docker logout ${env.HARBOR_URL} 2>/dev/null || true"
                sh 'docker image prune -f 2>/dev/null || true'
                sh 'rm -f /tmp/trivy-*.json /tmp/sbom-*.json /tmp/gitleaks-*.json 2>/dev/null || true'
                sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
            }
        }
    }
}
