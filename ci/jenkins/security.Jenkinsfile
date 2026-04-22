pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 180, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'ACTION',
            choices: ['INSTALL', 'UNINSTALL'],
            description: 'INSTALL — deploy selected security tools. UNINSTALL — remove selected.'
        )
        booleanParam(name: 'KEYCLOAK',         defaultValue: true,  description: 'Keycloak — IAM, SSO, OIDC provider')
        booleanParam(name: 'DEX',              defaultValue: true,  description: 'Dex — OIDC federation connector')
        booleanParam(name: 'AUTHENTIK',        defaultValue: true,  description: 'Authentik — identity provider alternative')
        booleanParam(name: 'ZITADEL',          defaultValue: true,  description: 'ZITADEL — cloud-native IAM')
        booleanParam(name: 'AUTHELIA',         defaultValue: true,  description: 'Authelia — SSO and 2FA portal')
        booleanParam(name: 'SPIRE',            defaultValue: true,  description: 'SPIFFE/SPIRE — workload identity attestation')
        booleanParam(name: 'POMERIUM',         defaultValue: true,  description: 'Pomerium — identity-aware access proxy')
        booleanParam(name: 'VAULT',            defaultValue: true,  description: 'HashiCorp Vault — secrets management and PKI')
        booleanParam(name: 'INFISICAL',        defaultValue: true,  description: 'Infisical — open-source secrets manager')
        booleanParam(name: 'OPA_GATEKEEPER',   defaultValue: true,  description: 'OPA Gatekeeper — policy-as-code admission control')
        booleanParam(name: 'KYVERNO',          defaultValue: true,  description: 'Kyverno — Kubernetes-native policy engine')
        booleanParam(name: 'KUBEWARDEN',       defaultValue: true,  description: 'Kubewarden — Wasm-based policy engine')
        booleanParam(name: 'OPENFGA',          defaultValue: true,  description: 'OpenFGA — relationship-based authorization')
        booleanParam(name: 'FALCO',            defaultValue: true,  description: 'Falco — runtime threat detection')
        booleanParam(name: 'TETRAGON',         defaultValue: true,  description: 'Tetragon — eBPF-based security enforcement')
        booleanParam(name: 'TRACEE',           defaultValue: true,  description: 'Tracee — eBPF runtime security and forensics')
        booleanParam(name: 'KUBEARMOR',        defaultValue: true,  description: 'KubeArmor — container-aware runtime security')
        booleanParam(name: 'CORAZA_WAF',       defaultValue: true,  description: 'Coraza WAF — OWASP ModSecurity-compatible WAF')
        booleanParam(name: 'CERT_MANAGER',     defaultValue: true,  description: 'cert-manager — TLS certificate automation')
        booleanParam(name: 'SONARQUBE',        defaultValue: true,  description: 'SonarQube — static code analysis and quality gates')
        booleanParam(name: 'TRIVY_OPERATOR',   defaultValue: true,  description: 'Trivy Operator — continuous container vulnerability scanning')
        booleanParam(name: 'CLAIR',            defaultValue: true,  description: 'Clair — container image vulnerability analysis')
        booleanParam(name: 'OPENVAS',          defaultValue: true,  description: 'OpenVAS — network vulnerability scanner')
        booleanParam(name: 'ANCHORE',          defaultValue: true,  description: 'Anchore — container image policy compliance')
        booleanParam(name: 'OWASP_ZAP',        defaultValue: true,  description: 'OWASP ZAP — DAST web application scanner')
        booleanParam(name: 'NUCLEI',           defaultValue: true,  description: 'Nuclei — CVE template-based scanner')
        booleanParam(name: 'KUBESCAPE',        defaultValue: true,  description: 'Kubescape — K8s security posture and compliance scanning')
        booleanParam(name: 'POLARIS',          defaultValue: true,  description: 'Polaris — Kubernetes workload best-practice validation')
        booleanParam(name: 'REKOR',            defaultValue: true,  description: 'Rekor — Sigstore transparency log')
        booleanParam(name: 'FULCIO',           defaultValue: true,  description: 'Fulcio — Sigstore certificate authority')
        booleanParam(name: 'NOTARY',           defaultValue: true,  description: 'Notary — container image signing and verification')
        booleanParam(name: 'SURICATA',         defaultValue: true,  description: 'Suricata — network IDS/IPS')
        booleanParam(name: 'ZEEK',             defaultValue: true,  description: 'Zeek — network traffic analysis')
        booleanParam(name: 'WAZUH',            defaultValue: true,  description: 'Wazuh — SIEM and XDR platform')
        booleanParam(name: 'DEPENDENCY_TRACK', defaultValue: true,  description: 'Dependency Track — SCA and SBOM analysis')
        booleanParam(name: 'DEFECTDOJO',       defaultValue: true,  description: 'DefectDojo — vulnerability management and deduplication')
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

        // ── Identity & Access ─────────────────────────────────────────────────

        stage('Keycloak') {
            when { expression { params.ACTION == 'INSTALL' && params.KEYCLOAK } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-keycloak.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-keycloak.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('keycloak')
                }
            }
        }

        stage('Dex') {
            when { expression { params.ACTION == 'INSTALL' && params.DEX } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-dex.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('dex')
                }
            }
        }

        stage('Authentik') {
            when { expression { params.ACTION == 'INSTALL' && params.AUTHENTIK } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-authentik.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-authentik.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('authentik')
                }
            }
        }

        stage('ZITADEL') {
            when { expression { params.ACTION == 'INSTALL' && params.ZITADEL } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-zitadel.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-zitadel.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zitadel')
                }
            }
        }

        stage('Authelia') {
            when { expression { params.ACTION == 'INSTALL' && params.AUTHELIA } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-authelia.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-authelia.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('authelia')
                }
            }
        }

        stage('SPIRE') {
            when { expression { params.ACTION == 'INSTALL' && params.SPIRE } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-spire.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('spire')
                }
            }
        }

        stage('Pomerium') {
            when { expression { params.ACTION == 'INSTALL' && params.POMERIUM } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-pomerium.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-pomerium.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pomerium')
                }
            }
        }

        // ── Secrets Management ────────────────────────────────────────────────

        stage('Vault') {
            when { expression { params.ACTION == 'INSTALL' && params.VAULT } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-vault.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-vault.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('vault')
                }
            }
        }

        stage('Infisical') {
            when { expression { params.ACTION == 'INSTALL' && params.INFISICAL } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-infisical.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-infisical.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('infisical')
                }
            }
        }

        // ── Policy Engines ────────────────────────────────────────────────────

        stage('OPA Gatekeeper') {
            when { expression { params.ACTION == 'INSTALL' && params.OPA_GATEKEEPER } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-opa.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('gatekeeper-system')
                }
            }
        }

        stage('Kyverno') {
            when { expression { params.ACTION == 'INSTALL' && params.KYVERNO } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-kyverno.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kyverno')
                }
            }
        }

        stage('Kubewarden') {
            when { expression { params.ACTION == 'INSTALL' && params.KUBEWARDEN } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-kubewarden.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kubewarden')
                }
            }
        }

        stage('OpenFGA') {
            when { expression { params.ACTION == 'INSTALL' && params.OPENFGA } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-openfga.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('openfga')
                }
            }
        }

        // ── Runtime Security ──────────────────────────────────────────────────

        stage('Falco') {
            when { expression { params.ACTION == 'INSTALL' && params.FALCO } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-falco.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('falco')
                }
            }
        }

        stage('Tetragon') {
            when { expression { params.ACTION == 'INSTALL' && params.TETRAGON } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-tetragon.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('tetragon')
                }
            }
        }

        stage('Tracee') {
            when { expression { params.ACTION == 'INSTALL' && params.TRACEE } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-tracee.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('tracee')
                }
            }
        }

        stage('KubeArmor') {
            when { expression { params.ACTION == 'INSTALL' && params.KUBEARMOR } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-kubearmor.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kubearmor')
                }
            }
        }

        // ── WAF & Certificates ────────────────────────────────────────────────

        stage('Coraza WAF') {
            when { expression { params.ACTION == 'INSTALL' && params.CORAZA_WAF } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-coraza-waf.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('coraza-waf')
                }
            }
        }

        stage('cert-manager') {
            when { expression { params.ACTION == 'INSTALL' && params.CERT_MANAGER } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-cert-manager.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('cert-manager')
                }
            }
        }

        // ── SAST Server ───────────────────────────────────────────────────────

        stage('SonarQube') {
            when { expression { params.ACTION == 'INSTALL' && params.SONARQUBE } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-sonarqube.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-sonarqube.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('sonarqube')
                }
            }
        }

        // ── Vulnerability Scanning ────────────────────────────────────────────

        stage('Trivy Operator') {
            when { expression { params.ACTION == 'INSTALL' && params.TRIVY_OPERATOR } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-trivy-operator.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('trivy-system')
                }
            }
        }

        stage('Clair') {
            when { expression { params.ACTION == 'INSTALL' && params.CLAIR } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-clair.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('clair')
                }
            }
        }

        stage('OpenVAS') {
            when { expression { params.ACTION == 'INSTALL' && params.OPENVAS } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-openvas.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-openvas.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('openvas')
                }
            }
        }

        stage('Anchore') {
            when { expression { params.ACTION == 'INSTALL' && params.ANCHORE } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-anchore.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-anchore.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('anchore')
                }
            }
        }

        // ── DAST ──────────────────────────────────────────────────────────────

        stage('OWASP ZAP') {
            when { expression { params.ACTION == 'INSTALL' && params.OWASP_ZAP } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-zap.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-zap.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zap')
                }
            }
        }

        stage('Nuclei') {
            when { expression { params.ACTION == 'INSTALL' && params.NUCLEI } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-nuclei.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-nuclei.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('nuclei')
                }
            }
        }

        // ── K8s Compliance ────────────────────────────────────────────────────

        stage('Kubescape') {
            when { expression { params.ACTION == 'INSTALL' && params.KUBESCAPE } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-kubescape.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kubescape')
                }
            }
        }

        stage('Polaris') {
            when { expression { params.ACTION == 'INSTALL' && params.POLARIS } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-polaris.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('polaris')
                }
            }
        }

        // ── Supply Chain ──────────────────────────────────────────────────────

        stage('Rekor') {
            when { expression { params.ACTION == 'INSTALL' && params.REKOR } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-rekor.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('rekor')
                }
            }
        }

        stage('Fulcio') {
            when { expression { params.ACTION == 'INSTALL' && params.FULCIO } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-fulcio.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('fulcio')
                }
            }
        }

        stage('Notary') {
            when { expression { params.ACTION == 'INSTALL' && params.NOTARY } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-notary.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('notary')
                }
            }
        }

        // ── Network Security ──────────────────────────────────────────────────

        stage('Suricata') {
            when { expression { params.ACTION == 'INSTALL' && params.SURICATA } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-suricata.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('suricata')
                }
            }
        }

        stage('Zeek') {
            when { expression { params.ACTION == 'INSTALL' && params.ZEEK } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-zeek.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zeek')
                }
            }
        }

        // ── SIEM / XDR ────────────────────────────────────────────────────────

        stage('Wazuh') {
            when { expression { params.ACTION == 'INSTALL' && params.WAZUH } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-wazuh.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-wazuh.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('wazuh')
                }
            }
        }

        // ── Vulnerability Management ──────────────────────────────────────────

        stage('Dependency Track') {
            when { expression { params.ACTION == 'INSTALL' && params.DEPENDENCY_TRACK } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-dependency-track.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-dependency-track.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('dependency-track')
                }
            }
        }

        stage('DefectDojo') {
            when { expression { params.ACTION == 'INSTALL' && params.DEFECTDOJO } }
            steps {
                script {
                    def s = load 'scripts/groovy/security-install-defectdojo.groovy'; s()
                    def c = load 'scripts/groovy/security-configure-defectdojo.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('defectdojo')
                }
            }
        }

        // ── CLI Tool Images (no K8s enhancements — not deployed to cluster) ───

        stage('Pull SAST CLI Images') {
            when { expression { params.ACTION == 'INSTALL' && params.DEFECTDOJO } }
            steps {
                sh 'docker pull python:3.13-slim'
                sh 'docker pull node:22-alpine'
                sh 'docker pull golangci/golangci-lint:latest'
                sh 'docker pull koalaman/shellcheck:stable'
                sh 'docker pull presidentbeef/brakeman:latest'
                sh 'docker pull semgrep/semgrep:latest'
                sh 'docker pull stoplight/spectral:latest'
                sh 'docker pull snyk/snyk:latest'
            }
        }

        stage('Pull Dependency Scanner Images') {
            when { expression { params.ACTION == 'INSTALL' && params.DEFECTDOJO } }
            steps {
                sh 'docker pull owasp/dependency-check:latest'
                sh 'docker pull aquasec/trivy:latest'
                sh 'docker pull anchore/grype:latest'
                sh 'docker pull anchore/syft:latest'
                sh 'docker pull docker/scout-cli:latest'
                sh 'docker pull fossas/fossa-cli:latest'
                sh 'docker pull vuls/vuls:latest'
                sh 'docker pull openscap/openscap:latest'
            }
        }

        stage('Pull Secret Scanner Images') {
            when { expression { params.ACTION == 'INSTALL' && params.DEFECTDOJO } }
            steps {
                sh 'docker pull gcr.io/projectsigstore/cosign:latest'
                sh 'docker pull gitguardian/ggshield:latest'
                sh 'docker pull zricethezav/gitleaks:latest'
                sh 'docker pull trufflesecurity/trufflehog:latest'
            }
        }

        stage('Pull IaC Scanner Images') {
            when { expression { params.ACTION == 'INSTALL' && params.DEFECTDOJO } }
            steps {
                sh 'docker pull checkmarx/kics:latest'
                sh 'docker pull tenable/terrascan:latest'
                sh 'docker pull aquasec/tfsec:latest'
                sh 'docker pull bridgecrew/checkov:latest'
            }
        }

        stage('Pull K8s Security CLI Images') {
            when { expression { params.ACTION == 'INSTALL' && params.DEFECTDOJO } }
            steps {
                sh 'docker pull aquasec/kube-bench:latest'
                sh 'docker pull aquasec/kube-hunter:latest'
                sh 'docker pull mondoo/cnspec:latest'
                sh 'docker pull shopify/kubeaudit:latest'
                sh 'docker pull philipssoftware/tern:latest'
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall DefectDojo') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DEFECTDOJO } }
            steps {
                sh '''
                    helm uninstall defectdojo -n defectdojo --ignore-not-found || true
                    kubectl delete namespace defectdojo --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Dependency Track') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DEPENDENCY_TRACK } }
            steps {
                sh '''
                    helm uninstall dependency-track -n dependency-track --ignore-not-found || true
                    kubectl delete namespace dependency-track --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Wazuh') {
            when { expression { params.ACTION == 'UNINSTALL' && params.WAZUH } }
            steps {
                sh '''
                    helm uninstall wazuh -n wazuh --ignore-not-found || true
                    kubectl delete namespace wazuh --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Zeek') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ZEEK } }
            steps {
                sh '''
                    helm uninstall zeek -n zeek --ignore-not-found || true
                    kubectl delete namespace zeek --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Suricata') {
            when { expression { params.ACTION == 'UNINSTALL' && params.SURICATA } }
            steps {
                sh '''
                    helm uninstall suricata -n suricata --ignore-not-found || true
                    kubectl delete namespace suricata --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Notary') {
            when { expression { params.ACTION == 'UNINSTALL' && params.NOTARY } }
            steps {
                sh '''
                    helm uninstall notary -n notary --ignore-not-found || true
                    kubectl delete namespace notary --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Fulcio') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FULCIO } }
            steps {
                sh '''
                    helm uninstall fulcio -n fulcio --ignore-not-found || true
                    kubectl delete namespace fulcio --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Rekor') {
            when { expression { params.ACTION == 'UNINSTALL' && params.REKOR } }
            steps {
                sh '''
                    helm uninstall rekor -n rekor --ignore-not-found || true
                    kubectl delete namespace rekor --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Polaris') {
            when { expression { params.ACTION == 'UNINSTALL' && params.POLARIS } }
            steps {
                sh '''
                    helm uninstall polaris -n polaris --ignore-not-found || true
                    kubectl delete namespace polaris --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kubescape') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KUBESCAPE } }
            steps {
                sh '''
                    helm uninstall kubescape -n kubescape --ignore-not-found || true
                    kubectl delete namespace kubescape --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Nuclei') {
            when { expression { params.ACTION == 'UNINSTALL' && params.NUCLEI } }
            steps {
                sh '''
                    helm uninstall nuclei -n nuclei --ignore-not-found || true
                    kubectl delete namespace nuclei --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OWASP ZAP') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OWASP_ZAP } }
            steps {
                sh '''
                    helm uninstall zap -n zap --ignore-not-found || true
                    kubectl delete namespace zap --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Anchore') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ANCHORE } }
            steps {
                sh '''
                    helm uninstall anchore -n anchore --ignore-not-found || true
                    kubectl delete namespace anchore --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OpenVAS') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPENVAS } }
            steps {
                sh '''
                    helm uninstall openvas -n openvas --ignore-not-found || true
                    kubectl delete namespace openvas --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Clair') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CLAIR } }
            steps {
                sh '''
                    helm uninstall clair -n clair --ignore-not-found || true
                    kubectl delete namespace clair --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Trivy Operator') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TRIVY_OPERATOR } }
            steps {
                sh '''
                    helm uninstall trivy-operator -n trivy-system --ignore-not-found || true
                    kubectl delete namespace trivy-system --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall SonarQube') {
            when { expression { params.ACTION == 'UNINSTALL' && params.SONARQUBE } }
            steps {
                sh '''
                    helm uninstall sonarqube -n sonarqube --ignore-not-found || true
                    kubectl delete pvc --all -n sonarqube --ignore-not-found || true
                    kubectl delete namespace sonarqube --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall cert-manager') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CERT_MANAGER } }
            steps {
                sh '''
                    helm uninstall cert-manager -n cert-manager --ignore-not-found || true
                    kubectl delete namespace cert-manager --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Coraza WAF') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CORAZA_WAF } }
            steps {
                sh '''
                    helm uninstall coraza-waf -n coraza-waf --ignore-not-found || true
                    kubectl delete namespace coraza-waf --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall KubeArmor') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KUBEARMOR } }
            steps {
                sh '''
                    helm uninstall kubearmor -n kubearmor --ignore-not-found || true
                    kubectl delete namespace kubearmor --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Tracee') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TRACEE } }
            steps {
                sh '''
                    helm uninstall tracee -n tracee --ignore-not-found || true
                    kubectl delete namespace tracee --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Tetragon') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TETRAGON } }
            steps {
                sh '''
                    helm uninstall tetragon -n tetragon --ignore-not-found || true
                    kubectl delete namespace tetragon --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Falco') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FALCO } }
            steps {
                sh '''
                    helm uninstall falco -n falco --ignore-not-found || true
                    kubectl delete namespace falco --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OpenFGA') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPENFGA } }
            steps {
                sh '''
                    helm uninstall openfga -n openfga --ignore-not-found || true
                    kubectl delete namespace openfga --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kubewarden') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KUBEWARDEN } }
            steps {
                sh '''
                    helm uninstall kubewarden -n kubewarden --ignore-not-found || true
                    kubectl delete namespace kubewarden --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kyverno') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KYVERNO } }
            steps {
                sh '''
                    helm uninstall kyverno -n kyverno --ignore-not-found || true
                    kubectl delete namespace kyverno --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OPA Gatekeeper') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPA_GATEKEEPER } }
            steps {
                sh '''
                    helm uninstall opa -n gatekeeper-system --ignore-not-found || true
                    kubectl delete namespace gatekeeper-system --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Infisical') {
            when { expression { params.ACTION == 'UNINSTALL' && params.INFISICAL } }
            steps {
                sh '''
                    helm uninstall infisical -n infisical --ignore-not-found || true
                    kubectl delete namespace infisical --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Vault') {
            when { expression { params.ACTION == 'UNINSTALL' && params.VAULT } }
            steps {
                sh '''
                    helm uninstall vault -n vault --ignore-not-found || true
                    kubectl delete pvc --all -n vault --ignore-not-found || true
                    kubectl delete namespace vault --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pomerium') {
            when { expression { params.ACTION == 'UNINSTALL' && params.POMERIUM } }
            steps {
                sh '''
                    helm uninstall pomerium -n pomerium --ignore-not-found || true
                    kubectl delete namespace pomerium --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall SPIRE') {
            when { expression { params.ACTION == 'UNINSTALL' && params.SPIRE } }
            steps {
                sh '''
                    helm uninstall spire -n spire --ignore-not-found || true
                    kubectl delete namespace spire --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Authelia') {
            when { expression { params.ACTION == 'UNINSTALL' && params.AUTHELIA } }
            steps {
                sh '''
                    helm uninstall authelia -n authelia --ignore-not-found || true
                    kubectl delete namespace authelia --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall ZITADEL') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ZITADEL } }
            steps {
                sh '''
                    helm uninstall zitadel -n zitadel --ignore-not-found || true
                    kubectl delete namespace zitadel --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Authentik') {
            when { expression { params.ACTION == 'UNINSTALL' && params.AUTHENTIK } }
            steps {
                sh '''
                    helm uninstall authentik -n authentik --ignore-not-found || true
                    kubectl delete namespace authentik --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Dex') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DEX } }
            steps {
                sh '''
                    helm uninstall dex -n dex --ignore-not-found || true
                    kubectl delete namespace dex --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Keycloak') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KEYCLOAK } }
            steps {
                sh '''
                    helm uninstall keycloak -n keycloak --ignore-not-found || true
                    kubectl delete pvc --all -n keycloak --ignore-not-found || true
                    kubectl delete namespace keycloak --ignore-not-found || true
                '''
            }
        }

        stage('Remove CLI Tool Images') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KEYCLOAK } }
            steps {
                sh 'docker rmi python:3.13-slim node:22-alpine golangci/golangci-lint:latest koalaman/shellcheck:stable presidentbeef/brakeman:latest semgrep/semgrep:latest stoplight/spectral:latest snyk/snyk:latest || true'
                sh 'docker rmi owasp/dependency-check:latest aquasec/trivy:latest anchore/grype:latest anchore/syft:latest docker/scout-cli:latest fossas/fossa-cli:latest vuls/vuls:latest openscap/openscap:latest || true'
                sh 'docker rmi gcr.io/projectsigstore/cosign:latest gitguardian/ggshield:latest zricethezav/gitleaks:latest trufflesecurity/trufflehog:latest || true'
                sh 'docker rmi checkmarx/kics:latest tenable/terrascan:latest aquasec/tfsec:latest bridgecrew/checkov:latest || true'
                sh 'docker rmi aquasec/kube-bench:latest aquasec/kube-hunter:latest mondoo/cnspec:latest shopify/kubeaudit:latest philipssoftware/tern:latest || true'
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
        }
        success { echo "${params.ACTION} completed successfully." }
        failure { echo "${params.ACTION} failed — check stage logs above." }
        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
            sh 'docker image prune -f 2>/dev/null || true'
        }
    }
}
