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
            choices: ['INSTALL', 'UNINSTALL', 'CONFIGURE'],
            description: 'Install or uninstall the selected GitOps tools on Kubernetes. CONFIGURE applies post-install setup (app-of-apps, repo connections, image update policies, bootstrap).'
        )

        // GitOps Continuous Delivery
        booleanParam(name: 'ARGOCD',                defaultValue: false, description: 'ArgoCD — GitOps continuous delivery for Kubernetes')
        booleanParam(name: 'ARGO_ROLLOUTS',         defaultValue: false, description: 'Argo Rollouts — advanced deployment strategies (canary, blue-green)')
        booleanParam(name: 'ARGO_WORKFLOWS',        defaultValue: false, description: 'Argo Workflows — Kubernetes-native workflow engine')
        booleanParam(name: 'ARGO_EVENTS',           defaultValue: false, description: 'Argo Events — event-driven workflow automation')
        booleanParam(name: 'ARGOCD_IMAGE_UPDATER',  defaultValue: false, description: 'ArgoCD Image Updater — auto-update container images in ArgoCD apps')
        booleanParam(name: 'FLUXCD',                defaultValue: false, description: 'Flux CD — GitOps toolkit for Kubernetes (CNCF)')
        booleanParam(name: 'FLAGGER',               defaultValue: false, description: 'Flagger — progressive delivery operator (canary, A/B, blue-green)')
        booleanParam(name: 'WEAVE_GITOPS',          defaultValue: false, description: 'Weave GitOps — GitOps dashboard and CLI for Flux')

        // Secrets Management
        booleanParam(name: 'SEALED_SECRETS',        defaultValue: false, description: 'Sealed Secrets — encrypt K8s secrets for safe GitOps storage')
        booleanParam(name: 'EXTERNAL_SECRETS',      defaultValue: false, description: 'External Secrets Operator — sync secrets from Vault, AWS SSM, etc.')

        // Multi-tenancy / Platform Engineering
        booleanParam(name: 'VCLUSTER',              defaultValue: false, description: 'vCluster — virtual Kubernetes clusters for tenant isolation')
        booleanParam(name: 'GIMLET',                defaultValue: false, description: 'Gimlet — developer platform built on top of GitOps')
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
            }
        }

        stage('Load Kubeconfig') {
            steps {
                script {
                    def kubeconfigContent = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    sh "echo '${kubeconfigContent}' | base64 -d > ${env.WORKSPACE}/kubeconfig"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        stage('Install GitOps Tools') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    if (params.ARGOCD)               { def s = load 'scripts/groovy/gitops-install-argocd.groovy';               s() }
                    if (params.ARGO_ROLLOUTS)        { def s = load 'scripts/groovy/gitops-install-argo-rollouts.groovy';        s() }
                    if (params.ARGO_WORKFLOWS)       { def s = load 'scripts/groovy/gitops-install-argo-workflows.groovy';       s() }
                    if (params.ARGO_EVENTS)          { def s = load 'scripts/groovy/gitops-install-argo-events.groovy';          s() }
                    if (params.ARGOCD_IMAGE_UPDATER) { def s = load 'scripts/groovy/gitops-install-argocd-image-updater.groovy'; s() }
                    if (params.FLUXCD)               { def s = load 'scripts/groovy/gitops-install-fluxcd.groovy';               s() }
                    if (params.FLAGGER)              { def s = load 'scripts/groovy/gitops-install-flagger.groovy';              s() }
                    if (params.WEAVE_GITOPS)         { def s = load 'scripts/groovy/gitops-install-weave-gitops.groovy';         s() }
                    if (params.SEALED_SECRETS)       { def s = load 'scripts/groovy/gitops-install-sealed-secrets.groovy';       s() }
                    if (params.EXTERNAL_SECRETS)     { def s = load 'scripts/groovy/gitops-install-external-secrets.groovy';     s() }
                    if (params.VCLUSTER)             { def s = load 'scripts/groovy/gitops-install-vcluster.groovy';             s() }
                    if (params.GIMLET)               { def s = load 'scripts/groovy/gitops-install-gimlet.groovy';               s() }
                }
            }
        }

        stage('Uninstall GitOps Tools') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.ARGOCD)               { sh 'helm uninstall argocd               -n argocd               --ignore-not-found || true' }
                    if (params.ARGO_ROLLOUTS)        { sh 'helm uninstall argo-rollouts        -n argo-rollouts        --ignore-not-found || true' }
                    if (params.ARGO_WORKFLOWS)       { sh 'helm uninstall argo-workflows       -n argo-workflows       --ignore-not-found || true' }
                    if (params.ARGO_EVENTS)          { sh 'helm uninstall argo-events          -n argo-events          --ignore-not-found || true' }
                    if (params.ARGOCD_IMAGE_UPDATER) { sh 'helm uninstall argocd-image-updater -n argocd-image-updater --ignore-not-found || true' }
                    if (params.FLUXCD)               { sh 'helm uninstall fluxcd               -n flux-system          --ignore-not-found || true' }
                    if (params.FLAGGER)              { sh 'helm uninstall flagger              -n flagger              --ignore-not-found || true' }
                    if (params.WEAVE_GITOPS)         { sh 'helm uninstall weave-gitops         -n weave-gitops         --ignore-not-found || true' }
                    if (params.SEALED_SECRETS)       { sh 'helm uninstall sealed-secrets       -n sealed-secrets       --ignore-not-found || true' }
                    if (params.EXTERNAL_SECRETS)     { sh 'helm uninstall external-secrets     -n external-secrets     --ignore-not-found || true' }
                    if (params.VCLUSTER)             { sh 'helm uninstall vcluster             -n vcluster             --ignore-not-found || true' }
                    if (params.GIMLET)               { sh 'helm uninstall gimlet               -n gimlet               --ignore-not-found || true' }
                }
            }
        }

        stage('Configure GitOps Tools') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps {
                script {
                    // ArgoCD — admin password, Git repo connection, app-of-apps
                    if (params.ARGOCD)               { def s = load 'scripts/groovy/gitops-configure-argocd.groovy';               s() }
                    // Argo Rollouts — default analysis templates
                    if (params.ARGO_ROLLOUTS)        { def s = load 'scripts/groovy/gitops-configure-argo-rollouts.groovy';        s() }
                    // Argo Workflows — default serviceaccount, artifact repo
                    if (params.ARGO_WORKFLOWS)       { def s = load 'scripts/groovy/gitops-configure-argo-workflows.groovy';       s() }
                    // Argo Events — event bus, event source
                    if (params.ARGO_EVENTS)          { def s = load 'scripts/groovy/gitops-configure-argo-events.groovy';          s() }
                    // ArgoCD Image Updater — registry credentials, update policy
                    if (params.ARGOCD_IMAGE_UPDATER) { def s = load 'scripts/groovy/gitops-configure-argocd-image-updater.groovy'; s() }
                    // Flux CD — bootstrap GitRepository, Kustomization
                    if (params.FLUXCD)               { def s = load 'scripts/groovy/gitops-configure-fluxcd.groovy';               s() }
                    // Flagger — canary analysis templates, Prometheus integration
                    if (params.FLAGGER)              { def s = load 'scripts/groovy/gitops-configure-flagger.groovy';              s() }
                    // External Secrets — ClusterSecretStore pointing to Vault
                    if (params.EXTERNAL_SECRETS)     { def s = load 'scripts/groovy/gitops-configure-external-secrets.groovy';     s() }
                    // Sealed Secrets — export controller public key to infra.env
                    if (params.SEALED_SECRETS)       { def s = load 'scripts/groovy/gitops-configure-sealed-secrets.groovy';       s() }
                }
            }
        }
    }

    post {
        success {
            echo "${params.ACTION} completed successfully."
        }
        failure {
            echo "${params.ACTION} failed — check stage logs above."
        }
        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
    }
}
