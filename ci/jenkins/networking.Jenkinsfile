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
            description: 'Install or uninstall the selected networking tools on Kubernetes. CONFIGURE applies post-install setup (mTLS policies, service entries, ingress routes, intentions).'
        )

        // CNI Plugins
        booleanParam(name: 'CILIUM',         defaultValue: false, description: 'Cilium — eBPF-based CNI with L7 network policies and observability (CNCF)')
        booleanParam(name: 'CALICO',         defaultValue: false, description: 'Calico — CNI with BGP routing and NetworkPolicy enforcement')
        booleanParam(name: 'FLANNEL',        defaultValue: false, description: 'Flannel — simple overlay CNI for Kubernetes pods')
        booleanParam(name: 'WEAVE_NET',      defaultValue: false, description: 'Weave Net — mesh CNI with auto-discovery and encryption')
        booleanParam(name: 'ANTREA',         defaultValue: false, description: 'Antrea — OVS-based CNI for Kubernetes (VMware)')

        // Ingress Controllers
        booleanParam(name: 'NGINX_INGRESS',   defaultValue: false, description: 'Nginx Ingress Controller — most widely used K8s ingress')
        booleanParam(name: 'TRAEFIK',         defaultValue: false, description: 'Traefik — cloud-native edge router and ingress controller')
        booleanParam(name: 'HAPROXY_INGRESS', defaultValue: false, description: 'HAProxy Ingress Controller — high-performance ingress')
        booleanParam(name: 'CONTOUR',         defaultValue: false, description: 'Contour — Envoy-based ingress controller with HTTPProxy CRD')
        booleanParam(name: 'KONG',            defaultValue: false, description: 'Kong — API gateway and ingress controller')

        // Service Mesh
        booleanParam(name: 'ISTIO',          defaultValue: false, description: 'Istio — service mesh with mTLS, traffic management, and telemetry')
        booleanParam(name: 'LINKERD',        defaultValue: false, description: 'Linkerd — lightweight service mesh for Kubernetes (CNCF)')
        booleanParam(name: 'CONSUL',         defaultValue: false, description: 'Consul — service discovery, health checking, and service mesh')

        // DNS
        booleanParam(name: 'EXTERNAL_DNS',   defaultValue: false, description: 'ExternalDNS — syncs K8s services to Route53 / CloudDNS / Azure DNS')
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

        stage('Install Networking Tools') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    // CNI
                    if (params.CILIUM)         { def s = load 'scripts/groovy/networking-install-cilium.groovy';         s() }
                    if (params.CALICO)         { def s = load 'scripts/groovy/networking-install-calico.groovy';         s() }
                    if (params.FLANNEL)        { def s = load 'scripts/groovy/networking-install-flannel.groovy';        s() }
                    if (params.WEAVE_NET)      { def s = load 'scripts/groovy/networking-install-weave-net.groovy';      s() }
                    if (params.ANTREA)         { def s = load 'scripts/groovy/networking-install-antrea.groovy';         s() }
                    // Ingress
                    if (params.NGINX_INGRESS)  { def s = load 'scripts/groovy/networking-install-nginx-ingress.groovy';  s() }
                    if (params.TRAEFIK)        { def s = load 'scripts/groovy/networking-install-traefik.groovy';        s() }
                    if (params.HAPROXY_INGRESS){ def s = load 'scripts/groovy/networking-install-haproxy-ingress.groovy';s() }
                    if (params.CONTOUR)        { def s = load 'scripts/groovy/networking-install-contour.groovy';        s() }
                    if (params.KONG)           { def s = load 'scripts/groovy/networking-install-kong.groovy';           s() }
                    // Service Mesh
                    if (params.ISTIO)          { def s = load 'scripts/groovy/networking-install-istio.groovy';          s() }
                    if (params.LINKERD)        { def s = load 'scripts/groovy/networking-install-linkerd.groovy';        s() }
                    if (params.CONSUL)         { def s = load 'scripts/groovy/networking-install-consul.groovy';         s() }
                    // DNS
                    if (params.EXTERNAL_DNS)   { def s = load 'scripts/groovy/networking-install-external-dns.groovy';  s() }
                }
            }
        }

        stage('Uninstall Networking Tools') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.CILIUM)         { sh 'helm uninstall cilium         -n cilium         --ignore-not-found || true' }
                    if (params.CALICO)         { sh 'helm uninstall calico         -n calico         --ignore-not-found || true' }
                    if (params.FLANNEL)        { sh 'helm uninstall flannel        -n flannel        --ignore-not-found || true' }
                    if (params.WEAVE_NET)      { sh 'helm uninstall weave-net      -n weave-net      --ignore-not-found || true' }
                    if (params.ANTREA)         { sh 'helm uninstall antrea         -n antrea         --ignore-not-found || true' }
                    if (params.NGINX_INGRESS)  { sh 'helm uninstall nginx-ingress  -n nginx-ingress  --ignore-not-found || true' }
                    if (params.TRAEFIK)        { sh 'helm uninstall traefik        -n traefik        --ignore-not-found || true' }
                    if (params.HAPROXY_INGRESS){ sh 'helm uninstall haproxy-ingress -n haproxy-ingress --ignore-not-found || true' }
                    if (params.CONTOUR)        { sh 'helm uninstall contour        -n contour        --ignore-not-found || true' }
                    if (params.KONG)           { sh 'helm uninstall kong           -n kong           --ignore-not-found || true' }
                    if (params.ISTIO)          { sh 'helm uninstall istio          -n istio          --ignore-not-found || true' }
                    if (params.LINKERD)        { sh 'helm uninstall linkerd        -n linkerd        --ignore-not-found || true' }
                    if (params.CONSUL)         { sh 'helm uninstall consul         -n consul         --ignore-not-found || true' }
                    if (params.EXTERNAL_DNS)   { sh 'helm uninstall external-dns   -n external-dns   --ignore-not-found || true' }
                }
            }
        }

        stage('Configure Networking Tools') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps {
                script {
                    // Service Mesh — post-install mTLS, traffic policies, sidecars
                    if (params.ISTIO)        { def s = load 'scripts/groovy/networking-configure-istio.groovy';        s() }
                    if (params.LINKERD)      { def s = load 'scripts/groovy/networking-configure-linkerd.groovy';      s() }
                    if (params.CONSUL)       { def s = load 'scripts/groovy/networking-configure-consul.groovy';       s() }
                    // Ingress — default IngressClass, TLS, rate-limit middleware
                    if (params.TRAEFIK)      { def s = load 'scripts/groovy/networking-configure-traefik.groovy';      s() }
                    if (params.NGINX_INGRESS){ def s = load 'scripts/groovy/networking-configure-nginx-ingress.groovy';s() }
                    if (params.KONG)         { def s = load 'scripts/groovy/networking-configure-kong.groovy';         s() }
                    // DNS
                    if (params.EXTERNAL_DNS) { def s = load 'scripts/groovy/networking-configure-external-dns.groovy'; s() }
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
