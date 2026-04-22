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
            choices: ['INSTALL', 'UNINSTALL', 'CONFIGURE'],
            description: 'INSTALL — deploy all networking tools on Kubernetes. UNINSTALL — remove all. CONFIGURE — post-install setup (mTLS policies, service entries, ingress routes, intentions).'
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
                    if (!fileExists('infra.env')) {
                        error "infra.env not found — run the k8s-infra pipeline first to provision a cluster"
                    }
                    def kubeconfigContent = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!kubeconfigContent) {
                        error "KUBECONFIG_CONTENT missing from infra.env — run the k8s-infra pipeline first"
                    }
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kubeconfigContent
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        // ── INSTALL ──────────────────────────────────────────────────────────

        stage('Install Cilium') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-cilium.groovy'; s() } }
        }

        stage('Install Calico') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-calico.groovy'; s() } }
        }

        stage('Install Flannel') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-flannel.groovy'; s() } }
        }

        stage('Install Weave Net') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-weave-net.groovy'; s() } }
        }

        stage('Install Antrea') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-antrea.groovy'; s() } }
        }

        stage('Install Nginx Ingress') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-nginx-ingress.groovy'; s() } }
        }

        stage('Install Traefik') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-traefik.groovy'; s() } }
        }

        stage('Install HAProxy Ingress') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-haproxy-ingress.groovy'; s() } }
        }

        stage('Install Contour') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-contour.groovy'; s() } }
        }

        stage('Install Kong') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-kong.groovy'; s() } }
        }

        stage('Install Istio') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-istio.groovy'; s() } }
        }

        stage('Install Linkerd') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-linkerd.groovy'; s() } }
        }

        stage('Install Consul') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-consul.groovy'; s() } }
        }

        stage('Install External DNS') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/networking-install-external-dns.groovy'; s() } }
        }

        // ── CONFIGURE ────────────────────────────────────────────────────────

        stage('Configure Istio') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-istio.groovy'; s() } }
        }

        stage('Configure Linkerd') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-linkerd.groovy'; s() } }
        }

        stage('Configure Consul') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-consul.groovy'; s() } }
        }

        stage('Configure Traefik') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-traefik.groovy'; s() } }
        }

        stage('Configure Nginx Ingress') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-nginx-ingress.groovy'; s() } }
        }

        stage('Configure Kong') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-kong.groovy'; s() } }
        }

        stage('Configure External DNS') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/networking-configure-external-dns.groovy'; s() } }
        }

        // ── UNINSTALL (reverse install order) ────────────────────────────────

        stage('Uninstall External DNS') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall external-dns -n external-dns --ignore-not-found || true' }
        }

        stage('Uninstall Consul') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall consul -n consul --ignore-not-found || true' }
        }

        stage('Uninstall Linkerd') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall linkerd -n linkerd --ignore-not-found || true' }
        }

        stage('Uninstall Istio') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall istio -n istio --ignore-not-found || true' }
        }

        stage('Uninstall Kong') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kong -n kong --ignore-not-found || true' }
        }

        stage('Uninstall Contour') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall contour -n contour --ignore-not-found || true' }
        }

        stage('Uninstall HAProxy Ingress') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall haproxy-ingress -n haproxy-ingress --ignore-not-found || true' }
        }

        stage('Uninstall Traefik') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall traefik -n traefik --ignore-not-found || true' }
        }

        stage('Uninstall Nginx Ingress') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall nginx-ingress -n nginx-ingress --ignore-not-found || true' }
        }

        stage('Uninstall Antrea') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall antrea -n antrea --ignore-not-found || true' }
        }

        stage('Uninstall Weave Net') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall weave-net -n weave-net --ignore-not-found || true' }
        }

        stage('Uninstall Flannel') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall flannel -n flannel --ignore-not-found || true' }
        }

        stage('Uninstall Calico') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall calico -n calico --ignore-not-found || true' }
        }

        stage('Uninstall Cilium') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall cilium -n cilium --ignore-not-found || true' }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
        }
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
