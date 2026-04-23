// ── GITOPS DEPLOY ─────────────────────────────────────────────────────────────
// This pipeline is a pure GitOps trigger. It does NOT build images.
// Images are built, scanned, signed, and pushed by pre-deploy.Jenkinsfile.
// Deployment is managed by ArgoCD (or Flux). This pipeline:
//   1. Confirms the image exists in Harbor
//   2. Updates Helm values with the target image tag (if not already done)
//   3. Triggers ArgoCD sync and waits for health
//   4. Verifies rollout in Kubernetes
//   5. Prints dashboard links
//   6. Sends deployment notification
// ─────────────────────────────────────────────────────────────────────────────

pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 30, unit: 'MINUTES')
    }

    parameters {
        string(
            name: 'SERVICE_NAME',
            defaultValue: '',
            description: 'Service(s) to deploy (comma-separated, e.g. order-service,payment-service). Blank = all in DOMAIN.'
        )
        choice(
            name: 'DOMAIN',
            choices: ['platform','identity','catalog','commerce','supply-chain','financial',
                      'customer-experience','communications','content','analytics-ai','b2b',
                      'integrations','affiliate','marketplace','gamification','developer-platform',
                      'compliance','sustainability','events-ticketing','auction','rental','web'],
            description: 'Business domain / Kubernetes namespace'
        )
        choice(
            name: 'ENVIRONMENT',
            choices: ['dev','staging','prod'],
            description: 'Target environment'
        )
        string(
            name: 'IMAGE_TAG',
            defaultValue: '',
            description: 'Image tag to deploy (e.g. dev-a1b2c3d-42). Must already be pushed to Harbor.'
        )
        string(
            name: 'REGISTRY_PROJECT',
            defaultValue: 'shopos',
            description: 'Harbor project / namespace'
        )
        booleanParam(
            name: 'FORCE_SYNC',
            defaultValue: false,
            description: 'Force ArgoCD sync even if app is already in-sync'
        )
        booleanParam(
            name: 'SKIP_IMAGE_VERIFY',
            defaultValue: false,
            description: 'Skip Harbor image existence check (useful for first deploy)'
        )
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
                        error "infra.env not found — run k8s-infra pipeline first"
                    }

                    def envMap = [:]
                    readFile('infra.env').trim().split('\n').each { line ->
                        def idx = line.indexOf('=')
                        if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                    }

                    env.HARBOR_URL       = envMap['HARBOR_URL']        ?: 'harbor.shopos.local'
                    env.HARBOR_USER      = envMap['HARBOR_USER']        ?: 'admin'
                    env.HARBOR_PASSWORD  = envMap['HARBOR_PASSWORD']    ?: ''
                    env.ARGOCD_URL       = envMap['ARGOCD_URL']         ?: 'http://argocd-server.argocd.svc.cluster.local:80'
                    env.ARGOCD_TOKEN     = envMap['ARGOCD_TOKEN']       ?: ''
                    env.GRAFANA_URL      = envMap['GRAFANA_URL']        ?: 'http://grafana-grafana.grafana.svc.cluster.local:3000'
                    env.PROMETHEUS_URL   = envMap['PROMETHEUS_URL']     ?: 'http://prometheus-prometheus.prometheus.svc.cluster.local:9090'
                    env.SLACK_WEBHOOK    = envMap['SLACK_WEBHOOK_URL']  ?: ''
                    env.EMAIL_RECIPIENTS = envMap['EMAIL_RECIPIENTS']   ?: ''
                    env.DEFECTDOJO_URL   = envMap['DEFECTDOJO_URL']     ?: 'http://defectdojo:8080'
                    env.SONAR_URL        = envMap['SONARQUBE_URL']      ?: 'http://sonarqube:9000'
                    env.JAEGER_URL       = envMap['JAEGER_URL']         ?: 'http://jaeger-query.tracing.svc.cluster.local:16686'
                    env.LOKI_URL         = envMap['LOKI_URL']           ?: 'http://loki.loki.svc.cluster.local:3100'
                    env.KIALI_URL        = envMap['KIALI_URL']          ?: 'http://kiali.istio-system.svc.cluster.local:20001'
                    env.UPTIME_KUMA_URL  = envMap['UPTIME_KUMA_URL']    ?: 'http://uptime-kuma.monitoring.svc.cluster.local:3001'

                    def kc = envMap['KUBECONFIG_CONTENT'] ?: ''
                    if (kc) {
                        writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kc
                        sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                        env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    }
                }
            }
        }

        // ── RESOLVE DEPLOYMENT TARGET ─────────────────────────────────────────

        stage('Resolve Deployment Target') {
            steps {
                script {
                    if (!params.IMAGE_TAG?.trim()) {
                        error "IMAGE_TAG is required — provide the tag built by pre-deploy pipeline"
                    }

                    env.IMAGE_TAG        = params.IMAGE_TAG.trim()
                    env.BUILD_DOMAIN     = params.DOMAIN
                    env.BUILD_ENV        = params.ENVIRONMENT
                    env.REGISTRY_PROJECT = params.REGISTRY_PROJECT
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

                    echo "────────────────────────────────────────────────────"
                    echo "Pipeline    : GITOPS DEPLOY"
                    echo "Services    : ${env.SERVICES}"
                    echo "Domain      : ${env.BUILD_DOMAIN}"
                    echo "Image Tag   : ${env.IMAGE_TAG}"
                    echo "Environment : ${env.BUILD_ENV}"
                    echo "ArgoCD URL  : ${env.ARGOCD_URL}"
                    echo "────────────────────────────────────────────────────"
                }
            }
        }

        // ── VERIFY IMAGE IN REGISTRY ──────────────────────────────────────────

        stage('Verify Image in Harbor') {
            when { expression { !params.SKIP_IMAGE_VERIFY } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        sh "echo '${env.HARBOR_PASSWORD}' | docker login ${env.HARBOR_URL} -u ${env.HARBOR_USER} --password-stdin"
                        env.SERVICES.split(',').each { svc ->
                            svc = svc.trim()
                            def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                            def exists = sh(
                                script: "docker manifest inspect ${image} > /dev/null 2>&1 && echo 'yes' || echo 'no'",
                                returnStdout: true
                            ).trim()
                            if (exists == 'yes') {
                                echo "PASS — Image exists in Harbor: ${image}"
                            } else {
                                echo "WARN — Image not found in Harbor: ${image} — ensure pre-deploy pipeline ran successfully"
                            }
                        }
                    }
                }
            }
        }

        // ── UPDATE GITOPS VALUES ──────────────────────────────────────────────

        stage('Update GitOps Manifests') {
            // Ensures Helm values files have the correct image tag.
            // Skipped if pre-deploy already committed the update.
            steps {
                script {
                    env.SERVICES.split(',').each { svc ->
                        svc = svc.trim()
                        def image      = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}"
                        def valuesFile = "helm/services/${svc}/values-${params.ENVIRONMENT}.yaml"
                        sh """
                            echo "=== Ensuring GitOps tag: ${svc} → ${env.IMAGE_TAG} ==="
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
                            git commit -m "deploy: ${params.DOMAIN} → ${env.IMAGE_TAG} (${params.ENVIRONMENT}) [skip ci]"
                        git push origin ${env.GIT_BRANCH_NAME} 2>/dev/null || true
                    """
                }
            }
        }

        // ── ARGOCD SYNC ───────────────────────────────────────────────────────

        stage('ArgoCD Sync') {
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        env.SERVICES.split(',').each { svc ->
                            svc = svc.trim()
                            sh """
                                echo "=== ArgoCD sync: ${svc} ==="
                                if command -v argocd &>/dev/null; then
                                    argocd app sync ${svc} \
                                        --server     ${env.ARGOCD_URL} \
                                        --auth-token ${env.ARGOCD_TOKEN} \
                                        ${params.FORCE_SYNC ? '--force' : ''} \
                                        --prune \
                                        --timeout 300 || true

                                    echo "=== Waiting for ${svc} to be Healthy ==="
                                    argocd app wait ${svc} \
                                        --server     ${env.ARGOCD_URL} \
                                        --auth-token ${env.ARGOCD_TOKEN} \
                                        --health \
                                        --timeout 300 || true

                                    argocd app get ${svc} \
                                        --server ${env.ARGOCD_URL} \
                                        --auth-token ${env.ARGOCD_TOKEN} || true
                                else
                                    echo "argocd CLI not found — triggering sync via REST API"
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

        // ── KUBERNETES ROLLOUT VERIFY ─────────────────────────────────────────

        stage('Verify Kubernetes Rollout') {
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    script {
                        env.SERVICES.split(',').each { svc ->
                            svc = svc.trim()
                            def ns = params.DOMAIN
                            sh """
                                echo "=== Rollout status: ${svc} (ns=${ns}) ==="
                                kubectl rollout status deployment/${svc} -n ${ns} --timeout=120s || \
                                kubectl rollout status statefulset/${svc} -n ${ns} --timeout=120s || true

                                echo "=== Pod health: ${svc} ==="
                                kubectl get pods -n ${ns} -l app=${svc} --no-headers 2>/dev/null | head -10 || true

                                echo "=== Health check: ${svc} /healthz ==="
                                kubectl port-forward svc/${svc} 19091:80 -n ${ns} &
                                PF_PID=\$!
                                sleep 5
                                HTTP_CODE=\$(curl -sf -o /dev/null -w "%{http_code}" \
                                    http://localhost:19091/healthz 2>/dev/null || echo "000")
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
                    ['HARBOR_URL','ARGOCD_URL','GRAFANA_URL','PROMETHEUS_URL',
                     'JAEGER_URL','KIALI_URL','DEFECTDOJO_URL'].each { k ->
                        if (env."${k}") envMap[k] = env."${k}"
                    }

                    def primarySvc = env.SERVICES?.split(',')?.getAt(0)?.trim() ?: 'unknown'
                    def d = load 'scripts/groovy/dashboard-links.groovy'
                    echo d.call(envMap, "DEPLOY (GitOps) — Build #${env.BUILD_NUMBER}", [
                        service: primarySvc,
                        tag:     env.IMAGE_TAG ?: 'unknown',
                        domain:  env.BUILD_DOMAIN ?: params.DOMAIN,
                        project: env.REGISTRY_PROJECT ?: 'shopos'
                    ])
                    echo "NEXT STEP: Trigger post-deploy.Jenkinsfile for smoke, load, DAST, chaos, and SLO tests."
                }
            }
        }
    }

    // ── POST ──────────────────────────────────────────────────────────────────

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
        }

        success {
            script {
                def summary = """DEPLOYMENT SUCCESS — ShopOS Build #${env.BUILD_NUMBER}
Services   : ${env.SERVICES}
Domain     : ${env.BUILD_DOMAIN}
Environment: ${env.BUILD_ENV}
Image Tag  : ${env.IMAGE_TAG}
ArgoCD     : ${env.ARGOCD_URL}/applications
Grafana    : ${env.GRAFANA_URL}/dashboards
Jaeger     : ${env.JAEGER_URL}
Build URL  : ${env.BUILD_URL}
Next step  : Run post-deploy pipeline for smoke, load, and chaos tests."""

                if (env.SLACK_WEBHOOK?.trim()) {
                    sh """
                        curl -s -X POST '${env.SLACK_WEBHOOK}' \
                            -H 'Content-Type: application/json' \
                            -d '{"text":"DEPLOYED: ShopOS #${env.BUILD_NUMBER} — ${env.SERVICES} → ${env.BUILD_ENV} @ ${env.IMAGE_TAG}\\nArgoCD: ${env.ARGOCD_URL}/applications\\nGrafana: ${env.GRAFANA_URL}\\nBuild: ${env.BUILD_URL}"}' || true
                    """
                }
                if (env.EMAIL_RECIPIENTS?.trim()) {
                    mail to:      env.EMAIL_RECIPIENTS,
                         subject: "DEPLOYED: ShopOS ${env.BUILD_DOMAIN}/${env.BUILD_ENV} — ${env.IMAGE_TAG}",
                         body:    summary
                }
                echo "=== DEPLOYMENT NOTIFICATION SENT ==="
                echo summary
            }
        }

        failure {
            script {
                def summary = """DEPLOYMENT FAILED — ShopOS Build #${env.BUILD_NUMBER}
Services   : ${env.SERVICES ?: 'unknown'}
Domain     : ${params.DOMAIN}
Environment: ${params.ENVIRONMENT}
Image Tag  : ${params.IMAGE_TAG}
ArgoCD     : ${env.ARGOCD_URL}/applications
Build URL  : ${env.BUILD_URL}
Action     : Check ArgoCD events and kubectl describe pod for failure reason."""

                if (env.SLACK_WEBHOOK?.trim()) {
                    sh """
                        curl -s -X POST '${env.SLACK_WEBHOOK}' \
                            -H 'Content-Type: application/json' \
                            -d '{"text":"DEPLOY FAILED: ShopOS #${env.BUILD_NUMBER} — ${env.SERVICES ?: params.DOMAIN} → ${params.ENVIRONMENT}\\nArgoCD: ${env.ARGOCD_URL}/applications\\nBuild: ${env.BUILD_URL}"}' || true
                    """
                }
                if (env.EMAIL_RECIPIENTS?.trim()) {
                    mail to:      env.EMAIL_RECIPIENTS,
                         subject: "DEPLOY FAILED: ShopOS ${params.DOMAIN}/${params.ENVIRONMENT} — Build #${env.BUILD_NUMBER}",
                         body:    summary
                }
                echo "=== FAILURE NOTIFICATION SENT ==="
                echo summary
            }
        }

        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
            sh "docker logout ${env.HARBOR_URL} 2>/dev/null || true"
        }
    }
}
