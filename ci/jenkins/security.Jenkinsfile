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
            choices: ['INSTALL', 'UNINSTALL', 'CONFIGURE'],
            description: 'INSTALL — deploy all security tools on Kubernetes and pull CLI tool images. CONFIGURE — post-install setup. UNINSTALL — remove all.'
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

        // ── INSTALL — Identity & Access ───────────────────────────────────────

        stage('Install Keycloak') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-keycloak.groovy'; s() } }
        }

        stage('Install Dex') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-dex.groovy'; s() } }
        }

        stage('Install Authentik') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-authentik.groovy'; s() } }
        }

        stage('Install ZITADEL') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-zitadel.groovy'; s() } }
        }

        stage('Install Authelia') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-authelia.groovy'; s() } }
        }

        stage('Install SPIRE') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-spire.groovy'; s() } }
        }

        stage('Install Pomerium') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-pomerium.groovy'; s() } }
        }

        // ── INSTALL — Secrets Management ──────────────────────────────────────

        stage('Install Vault') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-vault.groovy'; s() } }
        }

        stage('Install Infisical') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-infisical.groovy'; s() } }
        }

        // ── INSTALL — Policy Engines ──────────────────────────────────────────

        stage('Install OPA Gatekeeper') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-opa.groovy'; s() } }
        }

        stage('Install Kyverno') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-kyverno.groovy'; s() } }
        }

        stage('Install Kubewarden') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-kubewarden.groovy'; s() } }
        }

        stage('Install OpenFGA') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-openfga.groovy'; s() } }
        }

        // ── INSTALL — Runtime Security ────────────────────────────────────────

        stage('Install Falco') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-falco.groovy'; s() } }
        }

        stage('Install Tetragon') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-tetragon.groovy'; s() } }
        }

        stage('Install Tracee') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-tracee.groovy'; s() } }
        }

        stage('Install KubeArmor') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-kubearmor.groovy'; s() } }
        }

        // ── INSTALL — WAF & Certificates ─────────────────────────────────────

        stage('Install Coraza WAF') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-coraza-waf.groovy'; s() } }
        }

        stage('Install cert-manager') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-cert-manager.groovy'; s() } }
        }

        // ── INSTALL — SAST Server ─────────────────────────────────────────────

        stage('Install SonarQube') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-sonarqube.groovy'; s() } }
        }

        // ── INSTALL — Vulnerability Scanning ──────────────────────────────────

        stage('Install Trivy Operator') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-trivy-operator.groovy'; s() } }
        }

        stage('Install Clair') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-clair.groovy'; s() } }
        }

        stage('Install OpenVAS') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-openvas.groovy'; s() } }
        }

        stage('Install Anchore') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-anchore.groovy'; s() } }
        }

        // ── INSTALL — DAST ────────────────────────────────────────────────────

        stage('Install OWASP ZAP') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-zap.groovy'; s() } }
        }

        stage('Install Nuclei') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-nuclei.groovy'; s() } }
        }

        // ── INSTALL — K8s Compliance ──────────────────────────────────────────

        stage('Install Kubescape') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-kubescape.groovy'; s() } }
        }

        stage('Install Polaris') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-polaris.groovy'; s() } }
        }

        // ── INSTALL — Supply Chain ────────────────────────────────────────────

        stage('Install Rekor') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-rekor.groovy'; s() } }
        }

        stage('Install Fulcio') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-fulcio.groovy'; s() } }
        }

        stage('Install Notary') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-notary.groovy'; s() } }
        }

        // ── INSTALL — Network Security ────────────────────────────────────────

        stage('Install Suricata') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-suricata.groovy'; s() } }
        }

        stage('Install Zeek') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-zeek.groovy'; s() } }
        }

        // ── INSTALL — SIEM / XDR ──────────────────────────────────────────────

        stage('Install Wazuh') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-wazuh.groovy'; s() } }
        }

        // ── INSTALL — Vulnerability Management ───────────────────────────────

        stage('Install Dependency Track') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-dependency-track.groovy'; s() } }
        }

        stage('Install DefectDojo') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/security-install-defectdojo.groovy'; s() } }
        }

        // ── INSTALL — CLI Tool Images (docker pull) ───────────────────────────

        stage('Pull SAST CLI Images') {
            when { expression { params.ACTION == 'INSTALL' } }
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
            when { expression { params.ACTION == 'INSTALL' } }
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
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                sh 'docker pull gcr.io/projectsigstore/cosign:latest'
                sh 'docker pull gitguardian/ggshield:latest'
                sh 'docker pull zricethezav/gitleaks:latest'
                sh 'docker pull trufflesecurity/trufflehog:latest'
            }
        }

        stage('Pull IaC Scanner Images') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                sh 'docker pull checkmarx/kics:latest'
                sh 'docker pull tenable/terrascan:latest'
                sh 'docker pull aquasec/tfsec:latest'
                sh 'docker pull bridgecrew/checkov:latest'
            }
        }

        stage('Pull K8s Security CLI Images') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                sh 'docker pull aquasec/kube-bench:latest'
                sh 'docker pull aquasec/kube-hunter:latest'
                sh 'docker pull mondoo/cnspec:latest'
                sh 'docker pull shopify/kubeaudit:latest'
                sh 'docker pull philipssoftware/tern:latest'
            }
        }

        // ── CONFIGURE ────────────────────────────────────────────────────────

        stage('Configure Keycloak') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-keycloak.groovy'; s() } }
        }

        stage('Configure Authentik') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-authentik.groovy'; s() } }
        }

        stage('Configure ZITADEL') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-zitadel.groovy'; s() } }
        }

        stage('Configure Authelia') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-authelia.groovy'; s() } }
        }

        stage('Configure Pomerium') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-pomerium.groovy'; s() } }
        }

        stage('Configure Vault') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-vault.groovy'; s() } }
        }

        stage('Configure Infisical') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-infisical.groovy'; s() } }
        }

        stage('Configure SonarQube') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-sonarqube.groovy'; s() } }
        }

        stage('Configure Anchore') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-anchore.groovy'; s() } }
        }

        stage('Configure OpenVAS') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-openvas.groovy'; s() } }
        }

        stage('Configure OWASP ZAP') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-zap.groovy'; s() } }
        }

        stage('Configure Nuclei') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-nuclei.groovy'; s() } }
        }

        stage('Configure Wazuh') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-wazuh.groovy'; s() } }
        }

        stage('Configure Dependency Track') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-dependency-track.groovy'; s() } }
        }

        stage('Configure DefectDojo') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/security-configure-defectdojo.groovy'; s() } }
        }

        // ── UNINSTALL (reverse install order) ────────────────────────────────

        stage('Uninstall DefectDojo') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall defectdojo -n defectdojo --ignore-not-found || true' }
        }

        stage('Uninstall Dependency Track') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall dependency-track -n dependency-track --ignore-not-found || true' }
        }

        stage('Uninstall Wazuh') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall wazuh -n wazuh --ignore-not-found || true' }
        }

        stage('Uninstall Zeek') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall zeek -n zeek --ignore-not-found || true' }
        }

        stage('Uninstall Suricata') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall suricata -n suricata --ignore-not-found || true' }
        }

        stage('Uninstall Notary') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall notary -n notary --ignore-not-found || true' }
        }

        stage('Uninstall Fulcio') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall fulcio -n fulcio --ignore-not-found || true' }
        }

        stage('Uninstall Rekor') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall rekor -n rekor --ignore-not-found || true' }
        }

        stage('Uninstall Polaris') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall polaris -n polaris --ignore-not-found || true' }
        }

        stage('Uninstall Kubescape') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kubescape -n kubescape --ignore-not-found || true' }
        }

        stage('Uninstall Nuclei') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall nuclei -n nuclei --ignore-not-found || true' }
        }

        stage('Uninstall OWASP ZAP') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall zap -n zap --ignore-not-found || true' }
        }

        stage('Uninstall Anchore') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall anchore -n anchore --ignore-not-found || true' }
        }

        stage('Uninstall OpenVAS') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall openvas -n openvas --ignore-not-found || true' }
        }

        stage('Uninstall Clair') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall clair -n clair --ignore-not-found || true' }
        }

        stage('Uninstall Trivy Operator') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall trivy-operator -n trivy-operator --ignore-not-found || true' }
        }

        stage('Uninstall SonarQube') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall sonarqube -n sonarqube --ignore-not-found || true' }
        }

        stage('Uninstall cert-manager') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall cert-manager -n cert-manager --ignore-not-found || true' }
        }

        stage('Uninstall Coraza WAF') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall coraza-waf -n coraza-waf --ignore-not-found || true' }
        }

        stage('Uninstall KubeArmor') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kubearmor -n kubearmor --ignore-not-found || true' }
        }

        stage('Uninstall Tracee') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall tracee -n tracee --ignore-not-found || true' }
        }

        stage('Uninstall Tetragon') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall tetragon -n tetragon --ignore-not-found || true' }
        }

        stage('Uninstall Falco') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall falco -n falco --ignore-not-found || true' }
        }

        stage('Uninstall OpenFGA') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall openfga -n openfga --ignore-not-found || true' }
        }

        stage('Uninstall Kubewarden') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kubewarden -n kubewarden --ignore-not-found || true' }
        }

        stage('Uninstall Kyverno') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kyverno -n kyverno --ignore-not-found || true' }
        }

        stage('Uninstall OPA Gatekeeper') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall opa -n opa --ignore-not-found || true' }
        }

        stage('Uninstall Infisical') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall infisical -n infisical --ignore-not-found || true' }
        }

        stage('Uninstall Vault') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall vault -n vault --ignore-not-found || true' }
        }

        stage('Uninstall Pomerium') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall pomerium -n pomerium --ignore-not-found || true' }
        }

        stage('Uninstall SPIRE') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall spire -n spire --ignore-not-found || true' }
        }

        stage('Uninstall Authelia') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall authelia -n authelia --ignore-not-found || true' }
        }

        stage('Uninstall ZITADEL') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall zitadel -n zitadel --ignore-not-found || true' }
        }

        stage('Uninstall Authentik') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall authentik -n authentik --ignore-not-found || true' }
        }

        stage('Uninstall Dex') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall dex -n dex --ignore-not-found || true' }
        }

        stage('Uninstall Keycloak') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall keycloak -n keycloak --ignore-not-found || true' }
        }

        stage('Remove CLI Tool Images') {
            when { expression { params.ACTION == 'UNINSTALL' } }
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
        success {
            echo "${params.ACTION} completed successfully."
        }
        failure {
            echo "${params.ACTION} failed — check stage logs above."
        }
        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
            sh 'docker image prune -f 2>/dev/null || true'
        }
    }
}
