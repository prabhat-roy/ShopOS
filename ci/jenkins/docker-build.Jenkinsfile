pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '20'))
        timeout(time: 240, unit: 'MINUTES')
    }

    parameters {
        string(
            name: 'SERVICE_NAME',
            defaultValue: '',
            description: 'Single service to build (e.g. order-service). Leave blank to build all in DOMAIN (or all 230 if DOMAIN also blank).'
        )
        choice(
            name: 'DOMAIN',
            choices: ['','platform','identity','catalog','commerce','supply-chain','financial',
                      'customer-experience','communications','content','analytics-ai','b2b',
                      'integrations','affiliate','marketplace','gamification','developer-platform',
                      'compliance','sustainability','web'],
            description: 'Domain filter — empty = build ALL 230 services across every domain'
        )
        string(
            name: 'IMAGE_TAG',
            defaultValue: '',
            description: 'Docker image tag. Defaults to dev-<git-sha>-<build-number> if blank.'
        )
        string(
            name: 'REGISTRY',
            defaultValue: '',
            description: 'Harbor registry URL (e.g. harbor.shopos.internal). Falls back to infra.env HARBOR_URL.'
        )
        string(
            name: 'REGISTRY_PROJECT',
            defaultValue: 'shopos',
            description: 'Harbor project / namespace'
        )
        booleanParam(
            name: 'PUSH_IMAGE',
            defaultValue: true,
            description: 'Push built images to registry'
        )
        booleanParam(
            name: 'SCAN_IMAGE',
            defaultValue: true,
            description: 'Run Trivy image scan after each build (HIGH,CRITICAL)'
        )
    }

    stages {

        // ──────────────────────────────────────────────────────────────────────
        stage('Checkout') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Load Environment') {
            steps {
                script {
                    def envMap = [:]
                    if (fileExists('infra.env')) {
                        readFile('infra.env').trim().split('\n').each { line ->
                            def idx = line.indexOf('=')
                            if (idx > 0) envMap[line[0..<idx].trim()] = line[(idx+1)..-1].trim()
                        }
                    }

                    env.HARBOR_URL      = params.REGISTRY?.trim()          ?: envMap['HARBOR_URL']      ?: 'harbor.shopos.internal'
                    env.HARBOR_USER     = envMap['HARBOR_USER']             ?: 'admin'
                    env.HARBOR_PASSWORD = envMap['HARBOR_PASSWORD']         ?: ''

                    sh """
                        echo "=== Registry login: ${env.HARBOR_URL} ==="
                        echo '${env.HARBOR_PASSWORD}' | \
                            docker login ${env.HARBOR_URL} -u ${env.HARBOR_USER} --password-stdin || true
                    """
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Resolve Build Context') {
            steps {
                script {
                    def gitSha = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                    env.IMAGE_TAG        = params.IMAGE_TAG?.trim() ?: "dev-${gitSha}-${env.BUILD_NUMBER}"
                    env.REGISTRY_PROJECT = params.REGISTRY_PROJECT

                    def allDomains = ['platform','identity','catalog','commerce','supply-chain',
                                      'financial','customer-experience','communications','content',
                                      'analytics-ai','b2b','integrations','affiliate','marketplace',
                                      'gamification','developer-platform','compliance','sustainability',
                                      'web']

                    if (params.SERVICE_NAME?.trim()) {
                        // Single service — domain must be set
                        env.SERVICE_LIST = params.SERVICE_NAME.trim()
                        env.DOMAIN_LIST  = params.DOMAIN ?: 'platform'
                    } else if (params.DOMAIN?.trim()) {
                        // All services in one domain
                        def svcs = sh(
                            script: "ls src/${params.DOMAIN}/ 2>/dev/null | tr '\\n' ',' | sed 's/,\$//'",
                            returnStdout: true
                        ).trim()
                        env.SERVICE_LIST = svcs
                        env.DOMAIN_LIST  = params.DOMAIN
                    } else {
                        // All 230 services — build domain:service pairs
                        def pairs = []
                        allDomains.each { d ->
                            def svcs = sh(
                                script: "ls src/${d}/ 2>/dev/null || true",
                                returnStdout: true
                            ).trim().split('\n')
                            svcs.each { s ->
                                if (s?.trim()) pairs << "${d}:${s.trim()}"
                            }
                        }
                        env.DOMAIN_SVC_PAIRS = pairs.join(',')
                        env.BUILD_ALL        = 'true'
                    }

                    echo "────────────────────────────────────────────────"
                    echo "Tag      : ${env.IMAGE_TAG}"
                    echo "Registry : ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}"
                    echo "Push     : ${params.PUSH_IMAGE}"
                    echo "Scan     : ${params.SCAN_IMAGE}"
                    if (env.BUILD_ALL == 'true') {
                        echo "Scope    : ALL services"
                    } else {
                        echo "Scope    : ${env.DOMAIN_LIST} / ${env.SERVICE_LIST}"
                    }
                    echo "────────────────────────────────────────────────"
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Docker Build') {
            steps {
                script {
                    def buildPairs = []

                    if (env.BUILD_ALL == 'true') {
                        env.DOMAIN_SVC_PAIRS.split(',').each { pair ->
                            def parts = pair.split(':')
                            buildPairs << [domain: parts[0], svc: parts[1]]
                        }
                    } else {
                        def domain = env.DOMAIN_LIST
                        env.SERVICE_LIST.split(',').each { svc ->
                            buildPairs << [domain: domain, svc: svc.trim()]
                        }
                    }

                    buildPairs.each { entry ->
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${entry.svc}:${env.IMAGE_TAG}"
                        def ctxDir = entry.domain == 'web' ? "src/web/${entry.svc}/" : "src/${entry.domain}/${entry.svc}/"
                        sh """
                            echo "=== Building: ${image} ==="
                            docker build \
                                --label "git.commit=${env.IMAGE_TAG}" \
                                --label "build.number=${env.BUILD_NUMBER}" \
                                --label "domain=${entry.domain}" \
                                -t ${image} \
                                ${ctxDir}
                        """
                    }
                    // store for later stages
                    env.BUILD_PAIRS_JSON = buildPairs.collect { "${it.domain}:${it.svc}" }.join(',')
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Image Scan (Trivy)') {
            when { expression { params.SCAN_IMAGE } }
            steps {
                script {
                    sh 'mkdir -p reports/image-scan'
                    env.BUILD_PAIRS_JSON.split(',').each { pair ->
                        def parts  = pair.split(':')
                        def svc    = parts[1]
                        def image  = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh """
                            echo "=== Trivy scan: ${image} ==="
                            if command -v trivy &>/dev/null; then
                                trivy image --exit-code 0 --severity HIGH,CRITICAL \
                                    --format json \
                                    --output reports/image-scan/trivy-${svc}.json \
                                    ${image} || true
                            else
                                docker run --rm \
                                    -v /var/run/docker.sock:/var/run/docker.sock \
                                    aquasec/trivy:latest image \
                                    --exit-code 0 --severity HIGH,CRITICAL \
                                    --format json \
                                    --output /tmp/trivy-${svc}.json \
                                    ${image} || true
                            fi
                        """
                    }
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Docker Push') {
            when { expression { params.PUSH_IMAGE } }
            steps {
                script {
                    // Re-login to refresh token
                    sh "echo '${env.HARBOR_PASSWORD}' | docker login ${env.HARBOR_URL} -u ${env.HARBOR_USER} --password-stdin || true"

                    env.BUILD_PAIRS_JSON.split(',').each { pair ->
                        def parts = pair.split(':')
                        def svc   = parts[1]
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh """
                            echo "=== Pushing: ${image} ==="
                            docker push ${image}
                        """
                    }
                    echo "All images pushed to ${env.HARBOR_URL}/${env.REGISTRY_PROJECT}"
                }
            }
        }

        // ──────────────────────────────────────────────────────────────────────
        stage('Cleanup') {
            steps {
                script {
                    env.BUILD_PAIRS_JSON.split(',').each { pair ->
                        def parts = pair.split(':')
                        def svc   = parts[1]
                        def image = "${env.HARBOR_URL}/${env.REGISTRY_PROJECT}/${svc}:${env.IMAGE_TAG}"
                        sh "docker rmi ${image} 2>/dev/null || true"
                    }
                    sh 'docker image prune -f 2>/dev/null || true'
                    sh "docker logout ${env.HARBOR_URL} 2>/dev/null || true"
                    echo "Cleanup complete."
                }
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
            archiveArtifacts artifacts: 'reports/**', allowEmptyArchive: true
            echo "Build #${env.BUILD_NUMBER} done. Tag: ${env.IMAGE_TAG}"
        }
        success {
            echo "SUCCESS — images built and pushed @ ${env.IMAGE_TAG}"
        }
        failure {
            echo "FAILED — check stage logs."
        }
    }
}
