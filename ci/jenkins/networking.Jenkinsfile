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
            choices: ['INSTALL', 'UNINSTALL'],
            description: 'INSTALL — deploy selected networking tools. UNINSTALL — remove selected.'
        )
        booleanParam(name: 'CILIUM',          defaultValue: true,  description: 'Cilium — eBPF-based CNI with NetworkPolicy and observability')
        booleanParam(name: 'CALICO',          defaultValue: true,  description: 'Calico — BGP-based CNI with NetworkPolicy')
        booleanParam(name: 'FLANNEL',         defaultValue: false, description: 'Flannel — simple overlay CNI (disable when Cilium/Calico active)')
        booleanParam(name: 'WEAVE_NET',       defaultValue: false, description: 'Weave Net — mesh overlay CNI')
        booleanParam(name: 'ANTREA',          defaultValue: false, description: 'Antrea — OVS-based CNI')
        booleanParam(name: 'NGINX_INGRESS',   defaultValue: true,  description: 'Nginx Ingress — standard ingress controller')
        booleanParam(name: 'TRAEFIK',         defaultValue: true,  description: 'Traefik — edge router with automatic service discovery')
        booleanParam(name: 'HAPROXY_INGRESS', defaultValue: true,  description: 'HAProxy Ingress — high-performance load balancer')
        booleanParam(name: 'CONTOUR',         defaultValue: true,  description: 'Contour — Envoy-based ingress for Kubernetes')
        booleanParam(name: 'KONG',            defaultValue: true,  description: 'Kong — API gateway and ingress controller')
        booleanParam(name: 'ISTIO',           defaultValue: true,  description: 'Istio — full service mesh with mTLS, traffic management, observability')
        booleanParam(name: 'LINKERD',         defaultValue: true,  description: 'Linkerd — lightweight Rust-based service mesh')
        booleanParam(name: 'CONSUL',          defaultValue: true,  description: 'Consul — service discovery, health checking, K/V config')
        booleanParam(name: 'EXTERNAL_DNS',    defaultValue: true,  description: 'External DNS — sync Kubernetes services to external DNS providers')
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
                    env.CLOUD_PROVIDER = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('CLOUD_PROVIDER=') }?.split('=', 2)?.last() ?: 'GCP'
                }
            }
        }

        // ── INSTALL + CONFIGURE + K8s ENHANCEMENTS ───────────────────────────

        stage('Cilium') {
            when { expression { params.ACTION == 'INSTALL' && params.CILIUM } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-cilium.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('cilium')
                }
            }
        }

        stage('Calico') {
            when { expression { params.ACTION == 'INSTALL' && params.CALICO } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-calico.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('calico-system')
                }
            }
        }

        stage('Flannel') {
            when { expression { params.ACTION == 'INSTALL' && params.FLANNEL } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-flannel.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kube-flannel')
                }
            }
        }

        stage('Weave Net') {
            when { expression { params.ACTION == 'INSTALL' && params.WEAVE_NET } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-weave-net.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('weave-net')
                }
            }
        }

        stage('Antrea') {
            when { expression { params.ACTION == 'INSTALL' && params.ANTREA } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-antrea.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('antrea')
                }
            }
        }

        stage('Nginx Ingress') {
            when { expression { params.ACTION == 'INSTALL' && params.NGINX_INGRESS } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-nginx-ingress.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-nginx-ingress.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('nginx-ingress')
                }
            }
        }

        stage('Traefik') {
            when { expression { params.ACTION == 'INSTALL' && params.TRAEFIK } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-traefik.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-traefik.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('traefik')
                }
            }
        }

        stage('HAProxy Ingress') {
            when { expression { params.ACTION == 'INSTALL' && params.HAPROXY_INGRESS } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-haproxy-ingress.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('haproxy-ingress')
                }
            }
        }

        stage('Contour') {
            when { expression { params.ACTION == 'INSTALL' && params.CONTOUR } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-contour.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('projectcontour')
                }
            }
        }

        stage('Kong') {
            when { expression { params.ACTION == 'INSTALL' && params.KONG } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-kong.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-kong.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kong')
                }
            }
        }

        stage('Istio') {
            when { expression { params.ACTION == 'INSTALL' && params.ISTIO } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-istio.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-istio.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('istio-system')
                }
            }
        }

        stage('Linkerd') {
            when { expression { params.ACTION == 'INSTALL' && params.LINKERD } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-linkerd.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-linkerd.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('linkerd')
                }
            }
        }

        stage('Consul') {
            when { expression { params.ACTION == 'INSTALL' && params.CONSUL } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-consul.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-consul.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('consul')
                }
            }
        }

        stage('External DNS') {
            when { expression { params.ACTION == 'INSTALL' && params.EXTERNAL_DNS } }
            steps {
                script {
                    def s = load 'scripts/groovy/networking-install-external-dns.groovy'; s()
                    def c = load 'scripts/groovy/networking-configure-external-dns.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('external-dns')
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall External DNS') {
            when { expression { params.ACTION == 'UNINSTALL' && params.EXTERNAL_DNS } }
            steps {
                sh '''
                    helm uninstall external-dns -n external-dns --ignore-not-found || true
                    kubectl delete namespace external-dns --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Consul') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CONSUL } }
            steps {
                sh '''
                    helm uninstall consul -n consul --ignore-not-found || true
                    kubectl delete pvc --all -n consul --ignore-not-found || true
                    kubectl delete namespace consul --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Linkerd') {
            when { expression { params.ACTION == 'UNINSTALL' && params.LINKERD } }
            steps {
                sh '''
                    helm uninstall linkerd -n linkerd --ignore-not-found || true
                    kubectl delete namespace linkerd --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Istio') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ISTIO } }
            steps {
                sh '''
                    helm uninstall istio -n istio-system --ignore-not-found || true
                    kubectl delete namespace istio-system --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kong') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KONG } }
            steps {
                sh '''
                    helm uninstall kong -n kong --ignore-not-found || true
                    kubectl delete namespace kong --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Contour') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CONTOUR } }
            steps {
                sh '''
                    helm uninstall contour -n projectcontour --ignore-not-found || true
                    kubectl delete namespace projectcontour --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall HAProxy Ingress') {
            when { expression { params.ACTION == 'UNINSTALL' && params.HAPROXY_INGRESS } }
            steps {
                sh '''
                    helm uninstall haproxy-ingress -n haproxy-ingress --ignore-not-found || true
                    kubectl delete namespace haproxy-ingress --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Traefik') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TRAEFIK } }
            steps {
                sh '''
                    helm uninstall traefik -n traefik --ignore-not-found || true
                    kubectl delete namespace traefik --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Nginx Ingress') {
            when { expression { params.ACTION == 'UNINSTALL' && params.NGINX_INGRESS } }
            steps {
                sh '''
                    helm uninstall nginx-ingress -n nginx-ingress --ignore-not-found || true
                    kubectl delete namespace nginx-ingress --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Antrea') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ANTREA } }
            steps {
                sh '''
                    helm uninstall antrea -n antrea --ignore-not-found || true
                    kubectl delete namespace antrea --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Weave Net') {
            when { expression { params.ACTION == 'UNINSTALL' && params.WEAVE_NET } }
            steps {
                sh '''
                    helm uninstall weave-net -n weave-net --ignore-not-found || true
                    kubectl delete namespace weave-net --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Flannel') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FLANNEL } }
            steps {
                sh '''
                    helm uninstall flannel -n kube-flannel --ignore-not-found || true
                    kubectl delete namespace kube-flannel --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Calico') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CALICO } }
            steps {
                sh '''
                    helm uninstall calico -n calico-system --ignore-not-found || true
                    kubectl delete namespace calico-system --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Cilium') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CILIUM } }
            steps {
                sh '''
                    helm uninstall cilium -n cilium --ignore-not-found || true
                    kubectl delete namespace cilium --ignore-not-found || true
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
