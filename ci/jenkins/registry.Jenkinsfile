pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 90, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'ACTION',
            choices: ['INSTALL', 'UNINSTALL', 'CREATE_CLOUD_REGISTRY', 'DESTROY_CLOUD_REGISTRY'],
            description: 'INSTALL — deploy selected registry tools. UNINSTALL — remove selected. CREATE/DESTROY — manage cloud-native container registry (ECR/GCR/ACR).'
        )
        // ── Container registries ──────────────────────────────────────────────
        booleanParam(name: 'HARBOR',        defaultValue: true,  description: 'Harbor — enterprise container registry with security scanning')
        booleanParam(name: 'ZOT',           defaultValue: true,  description: 'Zot — OCI-native lightweight registry')
        booleanParam(name: 'DISTRIBUTION',  defaultValue: true,  description: 'Distribution — CNCF reference OCI registry')
        booleanParam(name: 'QUAY',          defaultValue: true,  description: 'Quay — Red Hat enterprise registry')
        booleanParam(name: 'KRAKEN',        defaultValue: true,  description: 'Kraken — P2P Docker registry for large-scale builds')
        booleanParam(name: 'DRAGONFLY',     defaultValue: true,  description: 'Dragonfly — intelligent P2P-based image distribution')
        // ── Universal artifact repositories ──────────────────────────────────
        booleanParam(name: 'NEXUS',         defaultValue: true,  description: 'Nexus — universal artifact repository (Maven, npm, PyPI, Go, Docker)')
        booleanParam(name: 'PULP',          defaultValue: true,  description: 'Pulp — on-prem package management (RPM, deb, Python, container)')
        // ── Git servers ───────────────────────────────────────────────────────
        booleanParam(name: 'GITEA',         defaultValue: true,  description: 'Gitea — lightweight self-hosted Git service')
        booleanParam(name: 'FORGEJO',       defaultValue: true,  description: 'Forgejo — community-driven Gitea fork')
        booleanParam(name: 'GOGS',          defaultValue: true,  description: 'Gogs — minimal self-hosted Git service')
        booleanParam(name: 'GITBUCKET',     defaultValue: true,  description: 'GitBucket — GitHub-compatible Git platform (JVM)')
        booleanParam(name: 'ONEDEV',        defaultValue: true,  description: 'OneDev — Git server with built-in CI/CD')
        booleanParam(name: 'GITLAB',        defaultValue: false, description: 'GitLab — full DevOps platform (heavy, disabled by default)')
        // ── Helm chart repositories ───────────────────────────────────────────
        booleanParam(name: 'CHARTMUSEUM',   defaultValue: true,  description: 'ChartMuseum — Helm chart repository server')
        booleanParam(name: 'TERRAREG',      defaultValue: true,  description: 'Terrareg — Terraform module registry')
        // ── Language package registries ───────────────────────────────────────
        booleanParam(name: 'VERDACCIO',     defaultValue: true,  description: 'Verdaccio — npm/yarn private registry proxy')
        booleanParam(name: 'CNPMJS',        defaultValue: true,  description: 'Cnpmjs — npm private registry (cnpm)')
        booleanParam(name: 'PYPISERVER',    defaultValue: true,  description: 'Pypiserver — minimal PyPI-compatible server')
        booleanParam(name: 'DEVPI',         defaultValue: true,  description: 'Devpi — full-featured PyPI server and proxy')
        booleanParam(name: 'QUETZ',         defaultValue: true,  description: 'Quetz — conda package server')
        booleanParam(name: 'ATHENS',        defaultValue: true,  description: 'Athens — Go module proxy')
        booleanParam(name: 'GOPROXY',       defaultValue: true,  description: 'Goproxy — simple Go module proxy')
        booleanParam(name: 'REPOSILITE',    defaultValue: true,  description: 'Reposilite — lightweight Maven/Gradle repository')
        booleanParam(name: 'BAGET',         defaultValue: true,  description: 'BaGet — NuGet-compatible package server')
        booleanParam(name: 'KELLNR',        defaultValue: true,  description: 'Kellnr — Rust crate registry')
        booleanParam(name: 'ALEXANDRIE',    defaultValue: true,  description: 'Alexandrie — Rust crate registry (alternative)')
        booleanParam(name: 'GEMINABOX',     defaultValue: true,  description: 'Geminabox — Ruby gem server')
        booleanParam(name: 'CONAN_SERVER',  defaultValue: true,  description: 'Conan Server — C/C++ package registry')
        booleanParam(name: 'APTLY',         defaultValue: true,  description: 'Aptly — Debian/Ubuntu apt repository manager')
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        stage('Detect Cloud') {
            when { expression { params.ACTION in ['CREATE_CLOUD_REGISTRY', 'DESTROY_CLOUD_REGISTRY'] } }
            steps {
                script {
                    def detectCloud = load 'scripts/groovy/k8s-detect-cloud.groovy'
                    detectCloud()
                }
            }
        }

        stage('Create Cloud Registry') {
            when { expression { params.ACTION == 'CREATE_CLOUD_REGISTRY' } }
            steps {
                script {
                    def cloud = env.CLOUD_PROVIDER
                    if (cloud == 'AWS')        { def s = load 'scripts/groovy/repo-create-ecr.groovy';              s() }
                    else if (cloud == 'GCP')   { def s = load 'scripts/groovy/repo-create-artifact-registry.groovy'; s() }
                    else if (cloud == 'AZURE') { def s = load 'scripts/groovy/repo-create-acr.groovy';              s() }
                    else                       { error "Unsupported cloud provider: ${cloud}" }
                }
            }
        }

        stage('Destroy Cloud Registry') {
            when { expression { params.ACTION == 'DESTROY_CLOUD_REGISTRY' } }
            steps {
                script {
                    def cloud = env.CLOUD_PROVIDER
                    if (cloud == 'AWS')        { def s = load 'scripts/groovy/repo-delete-ecr.groovy';              s() }
                    else if (cloud == 'GCP')   { def s = load 'scripts/groovy/repo-delete-artifact-registry.groovy'; s() }
                    else if (cloud == 'AZURE') { def s = load 'scripts/groovy/repo-delete-acr.groovy';              s() }
                    else                       { error "Unsupported cloud provider: ${cloud}" }
                }
            }
        }

        stage('Load Kubeconfig') {
            when { expression { params.ACTION in ['INSTALL', 'UNINSTALL'] } }
            steps {
                script {
                    if (!fileExists('infra.env')) error "infra.env not found — run the k8s-infra pipeline first"
                    def content = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!content) error "KUBECONFIG_CONTENT missing from infra.env"
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: content
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    env.CLOUD_PROVIDER = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: 'GCP'
                }
            }
        }

        // ── INSTALL + CONFIGURE + K8s ENHANCEMENTS ───────────────────────────

        stage('Harbor') {
            when { expression { params.ACTION == 'INSTALL' && params.HARBOR } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-harbor.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-harbor.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('harbor')
                }
            }
        }

        stage('Zot') {
            when { expression { params.ACTION == 'INSTALL' && params.ZOT } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-zot.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-zot.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zot')
                }
            }
        }

        stage('Distribution') {
            when { expression { params.ACTION == 'INSTALL' && params.DISTRIBUTION } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-distribution.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('distribution')
                }
            }
        }

        stage('Quay') {
            when { expression { params.ACTION == 'INSTALL' && params.QUAY } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-quay.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('quay')
                }
            }
        }

        stage('Kraken') {
            when { expression { params.ACTION == 'INSTALL' && params.KRAKEN } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-kraken.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kraken')
                }
            }
        }

        stage('Dragonfly') {
            when { expression { params.ACTION == 'INSTALL' && params.DRAGONFLY } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-dragonfly.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('dragonfly')
                }
            }
        }

        stage('Nexus') {
            when { expression { params.ACTION == 'INSTALL' && params.NEXUS } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-nexus.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-nexus.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('nexus')
                }
            }
        }

        stage('Pulp') {
            when { expression { params.ACTION == 'INSTALL' && params.PULP } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-pulp.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pulp')
                }
            }
        }

        stage('Gitea') {
            when { expression { params.ACTION == 'INSTALL' && params.GITEA } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-gitea.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-gitea.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('gitea')
                }
            }
        }

        stage('Forgejo') {
            when { expression { params.ACTION == 'INSTALL' && params.FORGEJO } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-forgejo.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-forgejo.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('forgejo')
                }
            }
        }

        stage('Gogs') {
            when { expression { params.ACTION == 'INSTALL' && params.GOGS } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-gogs.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('gogs')
                }
            }
        }

        stage('GitBucket') {
            when { expression { params.ACTION == 'INSTALL' && params.GITBUCKET } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-gitbucket.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('gitbucket')
                }
            }
        }

        stage('OneDev') {
            when { expression { params.ACTION == 'INSTALL' && params.ONEDEV } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-onedev.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('onedev')
                }
            }
        }

        stage('GitLab') {
            when { expression { params.ACTION == 'INSTALL' && params.GITLAB } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-gitlab.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-gitlab.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('gitlab')
                }
            }
        }

        stage('ChartMuseum') {
            when { expression { params.ACTION == 'INSTALL' && params.CHARTMUSEUM } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-chartmuseum.groovy'; s()
                    def c = load 'scripts/groovy/registry-configure-chartmuseum.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('chartmuseum')
                }
            }
        }

        stage('Terrareg') {
            when { expression { params.ACTION == 'INSTALL' && params.TERRAREG } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-terrareg.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('terrareg')
                }
            }
        }

        stage('Verdaccio') {
            when { expression { params.ACTION == 'INSTALL' && params.VERDACCIO } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-verdaccio.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('verdaccio')
                }
            }
        }

        stage('Cnpmjs') {
            when { expression { params.ACTION == 'INSTALL' && params.CNPMJS } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-cnpmjs.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('cnpmjs')
                }
            }
        }

        stage('Pypiserver') {
            when { expression { params.ACTION == 'INSTALL' && params.PYPISERVER } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-pypiserver.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pypiserver')
                }
            }
        }

        stage('Devpi') {
            when { expression { params.ACTION == 'INSTALL' && params.DEVPI } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-devpi.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('devpi')
                }
            }
        }

        stage('Quetz') {
            when { expression { params.ACTION == 'INSTALL' && params.QUETZ } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-quetz.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('quetz')
                }
            }
        }

        stage('Athens') {
            when { expression { params.ACTION == 'INSTALL' && params.ATHENS } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-athens.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('athens')
                }
            }
        }

        stage('Goproxy') {
            when { expression { params.ACTION == 'INSTALL' && params.GOPROXY } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-goproxy.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('goproxy')
                }
            }
        }

        stage('Reposilite') {
            when { expression { params.ACTION == 'INSTALL' && params.REPOSILITE } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-reposilite.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('reposilite')
                }
            }
        }

        stage('BaGet') {
            when { expression { params.ACTION == 'INSTALL' && params.BAGET } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-baget.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('baget')
                }
            }
        }

        stage('Kellnr') {
            when { expression { params.ACTION == 'INSTALL' && params.KELLNR } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-kellnr.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kellnr')
                }
            }
        }

        stage('Alexandrie') {
            when { expression { params.ACTION == 'INSTALL' && params.ALEXANDRIE } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-alexandrie.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('alexandrie')
                }
            }
        }

        stage('Geminabox') {
            when { expression { params.ACTION == 'INSTALL' && params.GEMINABOX } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-geminabox.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('geminabox')
                }
            }
        }

        stage('Conan Server') {
            when { expression { params.ACTION == 'INSTALL' && params.CONAN_SERVER } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-conan-server.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('conan-server')
                }
            }
        }

        stage('Aptly') {
            when { expression { params.ACTION == 'INSTALL' && params.APTLY } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-aptly.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('aptly')
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall Aptly') {
            when { expression { params.ACTION == 'UNINSTALL' && params.APTLY } }
            steps {
                sh '''
                    helm uninstall aptly -n aptly --ignore-not-found || true
                    kubectl delete pvc --all -n aptly --ignore-not-found || true
                    kubectl delete namespace aptly --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Conan Server') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CONAN_SERVER } }
            steps {
                sh '''
                    helm uninstall conan-server -n conan-server --ignore-not-found || true
                    kubectl delete namespace conan-server --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Geminabox') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GEMINABOX } }
            steps {
                sh '''
                    helm uninstall geminabox -n geminabox --ignore-not-found || true
                    kubectl delete namespace geminabox --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Alexandrie') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ALEXANDRIE } }
            steps {
                sh '''
                    helm uninstall alexandrie -n alexandrie --ignore-not-found || true
                    kubectl delete pvc --all -n alexandrie --ignore-not-found || true
                    kubectl delete namespace alexandrie --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kellnr') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KELLNR } }
            steps {
                sh '''
                    helm uninstall kellnr -n kellnr --ignore-not-found || true
                    kubectl delete pvc --all -n kellnr --ignore-not-found || true
                    kubectl delete namespace kellnr --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall BaGet') {
            when { expression { params.ACTION == 'UNINSTALL' && params.BAGET } }
            steps {
                sh '''
                    helm uninstall baget -n baget --ignore-not-found || true
                    kubectl delete pvc --all -n baget --ignore-not-found || true
                    kubectl delete namespace baget --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Reposilite') {
            when { expression { params.ACTION == 'UNINSTALL' && params.REPOSILITE } }
            steps {
                sh '''
                    helm uninstall reposilite -n reposilite --ignore-not-found || true
                    kubectl delete pvc --all -n reposilite --ignore-not-found || true
                    kubectl delete namespace reposilite --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Goproxy') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GOPROXY } }
            steps {
                sh '''
                    helm uninstall goproxy -n goproxy --ignore-not-found || true
                    kubectl delete namespace goproxy --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Athens') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ATHENS } }
            steps {
                sh '''
                    helm uninstall athens -n athens --ignore-not-found || true
                    kubectl delete pvc --all -n athens --ignore-not-found || true
                    kubectl delete namespace athens --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Quetz') {
            when { expression { params.ACTION == 'UNINSTALL' && params.QUETZ } }
            steps {
                sh '''
                    helm uninstall quetz -n quetz --ignore-not-found || true
                    kubectl delete namespace quetz --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Devpi') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DEVPI } }
            steps {
                sh '''
                    helm uninstall devpi -n devpi --ignore-not-found || true
                    kubectl delete pvc --all -n devpi --ignore-not-found || true
                    kubectl delete namespace devpi --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pypiserver') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PYPISERVER } }
            steps {
                sh '''
                    helm uninstall pypiserver -n pypiserver --ignore-not-found || true
                    kubectl delete pvc --all -n pypiserver --ignore-not-found || true
                    kubectl delete namespace pypiserver --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Cnpmjs') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CNPMJS } }
            steps {
                sh '''
                    helm uninstall cnpmjs -n cnpmjs --ignore-not-found || true
                    kubectl delete namespace cnpmjs --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Verdaccio') {
            when { expression { params.ACTION == 'UNINSTALL' && params.VERDACCIO } }
            steps {
                sh '''
                    helm uninstall verdaccio -n verdaccio --ignore-not-found || true
                    kubectl delete pvc --all -n verdaccio --ignore-not-found || true
                    kubectl delete namespace verdaccio --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Terrareg') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TERRAREG } }
            steps {
                sh '''
                    helm uninstall terrareg -n terrareg --ignore-not-found || true
                    kubectl delete namespace terrareg --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall ChartMuseum') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CHARTMUSEUM } }
            steps {
                sh '''
                    helm uninstall chartmuseum -n chartmuseum --ignore-not-found || true
                    kubectl delete namespace chartmuseum --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall GitLab') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GITLAB } }
            steps {
                sh '''
                    helm uninstall gitlab -n gitlab --ignore-not-found || true
                    kubectl delete pvc --all -n gitlab --ignore-not-found || true
                    kubectl delete namespace gitlab --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OneDev') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ONEDEV } }
            steps {
                sh '''
                    helm uninstall onedev -n onedev --ignore-not-found || true
                    kubectl delete pvc --all -n onedev --ignore-not-found || true
                    kubectl delete namespace onedev --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall GitBucket') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GITBUCKET } }
            steps {
                sh '''
                    helm uninstall gitbucket -n gitbucket --ignore-not-found || true
                    kubectl delete pvc --all -n gitbucket --ignore-not-found || true
                    kubectl delete namespace gitbucket --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Gogs') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GOGS } }
            steps {
                sh '''
                    helm uninstall gogs -n gogs --ignore-not-found || true
                    kubectl delete pvc --all -n gogs --ignore-not-found || true
                    kubectl delete namespace gogs --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Forgejo') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FORGEJO } }
            steps {
                sh '''
                    helm uninstall forgejo -n forgejo --ignore-not-found || true
                    kubectl delete pvc --all -n forgejo --ignore-not-found || true
                    kubectl delete namespace forgejo --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Gitea') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GITEA } }
            steps {
                sh '''
                    helm uninstall gitea -n gitea --ignore-not-found || true
                    kubectl delete pvc --all -n gitea --ignore-not-found || true
                    kubectl delete namespace gitea --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pulp') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PULP } }
            steps {
                sh '''
                    helm uninstall pulp -n pulp --ignore-not-found || true
                    kubectl delete pvc --all -n pulp --ignore-not-found || true
                    kubectl delete namespace pulp --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Nexus') {
            when { expression { params.ACTION == 'UNINSTALL' && params.NEXUS } }
            steps {
                sh '''
                    helm uninstall nexus -n nexus --ignore-not-found || true
                    kubectl delete pvc --all -n nexus --ignore-not-found || true
                    kubectl delete namespace nexus --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Dragonfly') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DRAGONFLY } }
            steps {
                sh '''
                    helm uninstall dragonfly -n dragonfly --ignore-not-found || true
                    kubectl delete pvc --all -n dragonfly --ignore-not-found || true
                    kubectl delete namespace dragonfly --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kraken') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KRAKEN } }
            steps {
                sh '''
                    helm uninstall kraken -n kraken --ignore-not-found || true
                    kubectl delete namespace kraken --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Quay') {
            when { expression { params.ACTION == 'UNINSTALL' && params.QUAY } }
            steps {
                sh '''
                    helm uninstall quay -n quay --ignore-not-found || true
                    kubectl delete pvc --all -n quay --ignore-not-found || true
                    kubectl delete namespace quay --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Distribution') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DISTRIBUTION } }
            steps {
                sh '''
                    helm uninstall distribution -n distribution --ignore-not-found || true
                    kubectl delete pvc --all -n distribution --ignore-not-found || true
                    kubectl delete namespace distribution --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Zot') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ZOT } }
            steps {
                sh '''
                    helm uninstall zot -n zot --ignore-not-found || true
                    kubectl delete pvc --all -n zot --ignore-not-found || true
                    kubectl delete namespace zot --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Harbor') {
            when { expression { params.ACTION == 'UNINSTALL' && params.HARBOR } }
            steps {
                sh '''
                    helm uninstall harbor -n harbor --ignore-not-found || true
                    kubectl delete pvc --all -n harbor --ignore-not-found || true
                    kubectl delete namespace harbor --ignore-not-found || true
                '''
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
        }
        success { echo "${params.ACTION} completed successfully." }
        failure { echo "${params.ACTION} failed — check stage logs above." }
        cleanup { sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true" }
    }
}
