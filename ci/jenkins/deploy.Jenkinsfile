pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 180, unit: 'MINUTES')
    }

    parameters {
        string(
            name: 'SERVICE_NAME',
            defaultValue: '',
            description: 'Service to build (e.g. order-service). Leave blank to build ALL services in DOMAIN.'
        )
        choice(
            name: 'DOMAIN',
            choices: ['platform','identity','catalog','commerce','supply-chain','financial',
                      'customer-experience','communications','content','analytics-ai','b2b',
                      'integrations','affiliate'],
            description: 'Business domain containing the service(s)'
        )
        choice(
            name: 'ENVIRONMENT',
            choices: ['dev','staging','prod'],
            description: 'Target Kubernetes environment'
        )
        string(
            name: 'IMAGE_TAG',
            defaultValue: '',
            description: 'Docker image tag. Defaults to <env>-<git-sha> if blank.'
        )
        string(
            name: 'REGISTRY_PROJECT',
            defaultValue: 'shopos',
            description: 'Harbor project / namespace'
        )

        // ── Stage skip flags ──────────────────────────────────────────────────
        booleanParam(name: 'SKIP_SECRETS',    defaultValue: false, description: 'Skip secret scanning  (Gitleaks, TruffleHog, GitGuardian)')
        booleanParam(name: 'SKIP_SAST',       defaultValue: false, description: 'Skip SAST             (Semgrep, SonarQube, language linters, Snyk code)')
        booleanParam(name: 'SKIP_SCA',        defaultValue: false, description: 'Skip SCA & SBOM       (Trivy FS, Grype, Syft, OWASP DC, Docker Scout, Snyk OSS)')
        booleanParam(name: 'SKIP_IAC',        defaultValue: false, description: 'Skip IaC scanning     (Checkov, KICS, tfsec, Terrascan, Polaris, Kubeaudit, cnspec)')
        booleanParam(name: 'SKIP_LICENSE',    defaultValue: false, description: 'Skip license checks   (FOSSA, Tern, license-checker, go-licenses, pip-licenses)')
        booleanParam(name: 'SKIP_IMAGE_SCAN', defaultValue: false, description: 'Skip image scanning   (Trivy, Grype, Anchore, Clair, Docker Scout, Syft)')
        booleanParam(name: 'SKIP_K8S_AUDIT',  defaultValue: false, description: 'Skip K8s audit        (kube-bench, kube-hunter, Kubescape, Kubeaudit, cnspec)')
        booleanParam(name: 'SKIP_DAST',       defaultValue: true,  description: 'Skip DAST             (OWASP ZAP, Nuclei) — requires running services')
        booleanParam(name: 'SKIP_DEPLOY',     defaultValue: false, description: 'Build & scan only — skip Kubernetes deployment')
    }

    stages {

        // ──────────────────────────────────────────────────────────────────────
        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Load Environment') {
            steps {
                script {
                    if (!fileExists('infra.env')) {
                        error "infra.env not found — run k8s-infra and registry pipelines first to provision the cluster"
                    }

                    // Parse infra.env once
                    def envMap = [:]
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }

                    // Registry
                    env.HARBOR_URL          = envMap['HARBOR_URL']                ?: 'harbor.shopos.local'
                    env.HARBOR_USER         = envMap['HARBOR_USER']                ?: 'admin'
                    env.HARBOR_PASSWORD     = envMap['HARBOR_PASSWORD']            ?: ''

                    // SAST / analysis servers
                    env.SONAR_URL           = envMap['SONARQUBE_URL']              ?: 'http://sonarqube-sonarqube.sonarqube.svc.cluster.local:9000'
                    env.SONAR_TOKEN         = envMap['SONARQUBE_TOKEN']            ?: ''
                    env.SNYK_TOKEN          = envMap['SNYK_TOKEN']                 ?: ''
                    env.FOSSA_API_KEY       = envMap['FOSSA_API_KEY']              ?: ''
                    env.GITGUARDIAN_API_KEY = envMap['GITGUARDIAN_API_KEY']        ?: ''

                    // Image scanning servers
                    env.ANCHORE_URL         = envMap['ANCHORE_URL']                ?: ''
                    env.ANCHORE_PASSWORD    = envMap['ANCHORE_PASSWORD']           ?: 'foobar'
                    env.CLAIR_URL           = envMap['CLAIR_URL']                  ?: ''

                    // DAST servers
                    env.ZAP_URL             = envMap['ZAP_URL']                    ?: ''
                    env.NUCLEI_URL          = envMap['NUCLEI_URL']                 ?: ''

                    // Vulnerability management
                    env.DEFECTDOJO_URL      = envMap['DEFECTDOJO_URL']             ?: 'http://defectdojo-defectdojo.defectdojo.svc.cluster.local:8080'
                    env.DEFECTDOJO_TOKEN    = envMap['DEFECTDOJO_TOKEN']           ?: ''
                    env.DEPTRACK_URL        = envMap['DEPENDENCY_TRACK_URL']       ?: 'http://dependency-track-dependency-track.dependency-track.svc.cluster.local:8080'
                    env.DEPTRACK_KEY        = envMap['DEPTRACK_API_KEY']           ?: ''

                    // SIEM
                    env.WAZUH_URL           = envMap['WAZUH_URL']                  ?: ''
                    env.WAZUH_TOKEN         = envMap['WAZUH_TOKEN']                ?: ''

                    // Kubeconfig
                    def kubeconfigContent = envMap['KUBECONFIG_CONTENT'] ?: ''
                    if (kubeconfigContent) {
                        writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kubeconfigContent
                        sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                        env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    }

                    sh 'mkdir -p reports/sast reports/sca reports/secrets reports/iac reports/image-scan reports/license reports/k8s-audit reports/dast'

                    // Login to container registry early — session persists for all subsequent docker calls
                    sh """
                        echo "=== Registry login: ${envMap['HARBOR_URL'] ?: 'harbor.shopos.local'} ==="
                        echo '${envMap['HARBOR_PASSWORD'] ?: ''}' | \
                            docker login ${envMap['HARBOR_URL'] ?: 'harbor.shopos.local'} \
                            -u ${envMap['HARBOR_USER'] ?: 'admin'} \
                            --password-stdin || true
                    """
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Resolve Build Context') {
            steps {
                script {
                    def gitSha = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    // Tag format: <env>-<gitsha>-<buildnumber>  e.g. dev-a1b2c3d-42
                    env.IMAGE_TAG        = params.IMAGE_TAG?.trim() ?: "${params.ENVIRONMENT}-${gitSha}-${env.BUILD_NUMBER}"
                    env.BUILD_DOMAIN     = params.DOMAIN
                    env.BUILD_ENV        = params.ENVIRONMENT
                    env.REGISTRY_PROJECT = params.REGISTRY_PROJECT

                    if (params.SERVICE_NAME?.trim()) {
                        env.SERVICES = params.SERVICE_NAME.trim()
                    } else {
                        def svcList = sh(
                            script: "ls src/${params.DOMAIN}/ 2>/dev/null | tr '\\n' ','",
                            returnStdout: true
                        ).trim().replaceAll(/,$/, '')
                        env.SERVICES = svcList
                    }

                    def primarySvc = env.SERVICES.split(',')[0].trim()
                    def detector   = load 'scripts/groovy/deploy-language-detect.groovy'
                    env.LANGUAGE   = detector.call(primarySvc)

                    echo "────────────────────────────────────────────────"
                    echo "Services   : ${env.SERVICES}"
                    echo "Domain     : ${env.BUILD_DOMAIN}"
                    echo "Language   : ${env.LANGUAGE}"
                    echo "Tag        : ${env.IMAGE_TAG}"
                    echo "Environment: ${env.BUILD_ENV}"
                    echo "Registry   : ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}"
                    echo "────────────────────────────────────────────────"
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Pre-flight Checks') {
            // Verify all cluster infrastructure is ready before deploying services.
            // Blocking: Cilium, Traefik, Istio, cert-manager, Vault (unsealed), ESO, Kafka, Schema Registry, Harbor.
            // Non-blocking warnings: Keycloak, Kyverno, Prometheus.
            when { expression { !params.SKIP_DEPLOY } }
            steps {
                script {
                    def preflight = load 'scripts/groovy/deploy-preflight.groovy'
                    preflight.call()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Secret Scanning') {
            // Tools: Gitleaks · TruffleHog · GitGuardian ggshield · detect-secrets
            when { expression { !params.SKIP_SECRETS } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-secrets.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('SAST') {
            // Tools: Semgrep · ShellCheck · Spectral · SonarQube · Snyk Code
            //        GoSec · GolangCI · SpotBugs · PMD · Detekt
            //        Bandit · Pylint · Flake8 · Pyflakes
            //        ESLint · retire.js · Roslyn · cargo clippy · Scalastyle · Brakeman
            when { expression { !params.SKIP_SAST } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-sast.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('SCA & SBOM') {
            // Tools: Trivy FS · Grype FS · Syft (SPDX+CycloneDX) · OWASP DC
            //        Snyk OSS · Docker Scout FS · Vuls · OpenSCAP
            //        govulncheck · npm audit · pip-audit · cargo audit · Maven OWASP DC
            when { expression { !params.SKIP_SCA } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-sca.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('IaC & Manifest Scanning') {
            // Tools: Checkov · KICS · tfsec · Terrascan · Polaris · Kubeaudit · cnspec
            when { expression { !params.SKIP_IAC } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-iac.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('License Compliance') {
            // Tools: FOSSA · Tern · license-checker · go-licenses · pip-licenses
            when { expression { !params.SKIP_LICENSE } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-license.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Compile & Test') {
            // go build · mvn package · gradle build · pip install · npm ci
            // dotnet build · cargo build · sbt assembly
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-compile.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Docker Build') {
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh """
                            echo "=== Building: ${image} ==="
                            docker build \
                                --label "git.commit=${env.IMAGE_TAG}" \
                                --label "build.number=${env.BUILD_NUMBER}" \
                                --label "domain=${env.BUILD_DOMAIN}" \
                                --label "environment=${env.BUILD_ENV}" \
                                -t ${image} \
                                src/${env.BUILD_DOMAIN}/${svc}/
                        """
                    }
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Container Image Scan') {
            // Tools: Trivy image · Trivy secrets · Grype · Anchore Engine (K8s)
            //        Clair (K8s) · Docker Scout · Syft image SBOM
            when { expression { !params.SKIP_IMAGE_SCAN } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-image-scan.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Tag & Push to Registry') {
            steps {
                script {
                    // Re-login before push to ensure token hasn't expired during long scan stages
                    sh "echo '${env.HARBOR_PASSWORD}' | docker login ${env.HARBOR_URL} -u ${env.HARBOR_USER} --password-stdin"

                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        // Image name includes build number: <env>-<gitsha>-<buildnumber>
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh """
                            echo "=== Pushing: ${image} ==="
                            docker push ${image}
                        """
                        echo "Pushed ${image}"
                    }
                    echo "All images pushed to ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}"
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Supply Chain Signing') {
            // Tools: Cosign (keyless) · Notary/notation · Rekor transparency log
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"

                        // Cosign — keyless signing via OIDC
                        sh """
                            if command -v cosign &>/dev/null; then
                                echo "=== Signing: cosign — ${svc} ==="
                                cosign sign --yes ${image} || true
                            else
                                docker run --rm \
                                    -v /var/run/docker.sock:/var/run/docker.sock \
                                    gcr.io/projectsigstore/cosign:latest sign --yes ${image} || true
                            fi
                        """

                        // Notation (Notary v2)
                        sh """
                            if command -v notation &>/dev/null; then
                                echo "=== Signing: notation — ${svc} ==="
                                notation sign ${image} || true
                            fi
                        """
                    }
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Kubernetes Security Audit') {
            // Tools: kube-bench (CIS) · kube-hunter · Kubescape (NSA/MITRE)
            //        Kubeaudit (live) · cnspec
            when { expression { !params.SKIP_K8S_AUDIT } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-k8s-audit.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Deploy to Kubernetes') {
            when { expression { !params.SKIP_DEPLOY } }
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def ns         = params.DOMAIN
                        def imageRepo  = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}"
                        def valuesFile = "helm/charts/${svc}/values-${params.ENVIRONMENT}.yaml"

                        sh """
                            echo "=== Deploying: ${svc} → namespace=${ns} (${params.ENVIRONMENT}) ==="
                            helm upgrade --install ${svc} helm/charts/${svc} \
                                --namespace ${ns} \
                                --create-namespace \
                                --set image.repository=${imageRepo} \
                                --set image.tag=${env.IMAGE_TAG} \
                                --set environment=${params.ENVIRONMENT} \
                                \$([ -f ${valuesFile} ] && echo "-f ${valuesFile}" || true) \
                                --wait \
                                --timeout 5m
                        """
                    }
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Post-Deploy Health Check') {
            when { expression { !params.SKIP_DEPLOY } }
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def ns = params.DOMAIN

                        sh """
                            echo "=== Health check: ${svc} (namespace=${ns}) ==="
                            kubectl rollout status deployment/${svc} -n ${ns} --timeout=120s || true

                            kubectl port-forward svc/${svc} 19090:80 -n ${ns} &
                            PF_PID=\$!
                            sleep 5
                            HTTP_CODE=\$(curl -sf -o /dev/null -w "%{http_code}" \
                                http://localhost:19090/healthz 2>/dev/null || echo "000")
                            kill \$PF_PID 2>/dev/null || true
                            wait \$PF_PID 2>/dev/null || true

                            if [ "\$HTTP_CODE" = "200" ]; then
                                echo "PASS — ${svc} /healthz returned 200"
                            else
                                echo "WARN — ${svc} /healthz returned \$HTTP_CODE"
                            fi
                        """
                    }
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('DAST') {
            // Tools: OWASP ZAP (spider + active scan) · Nuclei (template-based)
            // Runs AFTER deploy so services are reachable inside the cluster.
            when { expression { !params.SKIP_DAST && !params.SKIP_DEPLOY } }
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-dast.groovy'
                    s()
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Security Report Upload') {
            // DefectDojo — all scan types
            // Dependency Track — CycloneDX SBOMs (source + image)
            // Wazuh — pipeline completion event
            steps {
                script {
                    def s = load 'scripts/groovy/deploy-report.groovy'
                    s()
                }
            }
        }
    }

    // ──────────────────────────────────────────────────────────────────────────
    post {
        always {
            // Archive all security reports
            archiveArtifacts artifacts: 'reports/**', allowEmptyArchive: true

            echo "Build #${env.BUILD_NUMBER} done — reports archived."
        }

        success {
            echo "SUCCESS — ${env.SERVICES} → ${params.ENVIRONMENT} @ ${env.IMAGE_TAG}"
        }

        failure {
            echo "FAILED — check stage logs. All partial reports archived for triage."
        }

        cleanup {
            script {
                echo "=== Cleanup ==="

                // Remove built images from Jenkins agent to free disk
                env.SERVICES?.split(',')?.each { svc ->
                    svc = svc?.trim()
                    if (svc && env.HARBOR_URL && env.REGISTRY_PROJECT && env.IMAGE_TAG) {
                        // Remove only the versioned image (no latest tag is used)
                        sh "docker rmi ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG} 2>/dev/null || true"
                    }
                }

                // Docker logout
                sh "docker logout ${env.HARBOR_URL} 2>/dev/null || true"

                // Remove dangling layers left by scanner containers
                sh 'docker image prune -f 2>/dev/null || true'

                // Remove scanner tool temp files
                sh 'rm -f /tmp/trivy-*.json /tmp/trivy-secret-*.json /tmp/sbom-img-*.json /tmp/nuclei-*.json /tmp/kubescape.json 2>/dev/null || true'

                // Remove kubeconfig written to workspace
                sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"

                echo "Cleanup complete."
            }
        }
    }
}
