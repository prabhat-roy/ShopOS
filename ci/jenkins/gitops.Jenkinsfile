pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 60, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'ACTION',
            choices: ['INSTALL', 'UNINSTALL'],
            description: 'INSTALL — deploy, configure and apply K8s enhancements for all GitOps tools. UNINSTALL — remove all.'
        )
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        stage('Load Kubeconfig') {
            steps {
                script {
                    if (!fileExists('infra.env')) error "infra.env not found — run the k8s-infra pipeline first"
                    def content = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!content) error "KUBECONFIG_CONTENT missing from infra.env"
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: content
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        // ── INSTALL + CONFIGURE + K8s ENHANCEMENTS ───────────────────────────

        stage('ArgoCD') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-argocd.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-argocd.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('argocd')
                }
            }
        }

        stage('Argo Rollouts') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-argo-rollouts.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-argo-rollouts.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('argo-rollouts')
                }
            }
        }

        stage('Argo Workflows') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-argo-workflows.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-argo-workflows.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('argo-workflows')
                }
            }
        }

        stage('Argo Events') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-argo-events.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-argo-events.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('argo-events')
                }
            }
        }

        stage('ArgoCD Image Updater') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-argocd-image-updater.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-argocd-image-updater.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('argocd-image-updater')
                }
            }
        }

        stage('Flux CD') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-fluxcd.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-fluxcd.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('flux-system')
                }
            }
        }

        stage('Flagger') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-flagger.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-flagger.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('flagger')
                }
            }
        }

        stage('Weave GitOps') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-weave-gitops.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-weave-gitops.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('weave-gitops')
                }
            }
        }

        stage('Sealed Secrets') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-sealed-secrets.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-sealed-secrets.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('sealed-secrets')
                }
            }
        }

        stage('External Secrets') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-external-secrets.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-external-secrets.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('external-secrets')
                }
            }
        }

        stage('vCluster') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-vcluster.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-vcluster.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('vcluster')
                }
            }
        }

        stage('Gimlet') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/gitops-install-gimlet.groovy'; s()
                    def c = load 'scripts/groovy/gitops-configure-gimlet.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('gimlet')
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall Gimlet') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall gimlet -n gimlet --ignore-not-found || true' }
        }

        stage('Uninstall vCluster') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall vcluster -n vcluster --ignore-not-found || true' }
        }

        stage('Uninstall External Secrets') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall external-secrets -n external-secrets --ignore-not-found || true' }
        }

        stage('Uninstall Sealed Secrets') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall sealed-secrets -n sealed-secrets --ignore-not-found || true' }
        }

        stage('Uninstall Weave GitOps') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall weave-gitops -n weave-gitops --ignore-not-found || true' }
        }

        stage('Uninstall Flagger') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall flagger -n flagger --ignore-not-found || true' }
        }

        stage('Uninstall Flux CD') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall fluxcd -n flux-system --ignore-not-found || true' }
        }

        stage('Uninstall ArgoCD Image Updater') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall argocd-image-updater -n argocd-image-updater --ignore-not-found || true' }
        }

        stage('Uninstall Argo Events') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall argo-events -n argo-events --ignore-not-found || true' }
        }

        stage('Uninstall Argo Workflows') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall argo-workflows -n argo-workflows --ignore-not-found || true' }
        }

        stage('Uninstall Argo Rollouts') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall argo-rollouts -n argo-rollouts --ignore-not-found || true' }
        }

        stage('Uninstall ArgoCD') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall argocd -n argocd --ignore-not-found || true' }
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
