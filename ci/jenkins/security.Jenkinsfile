pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '10'))
        timeout(time: 120, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'ACTION',
            choices: ['INSTALL', 'UNINSTALL', 'CONFIGURE'],
            description: 'INSTALL — deploy K8s tools via Helm + pull CLI tool images. CONFIGURE — post-install setup. UNINSTALL — remove all.'
        )

        // ── Identity & Access (K8s) ───────────────────────────────────────
        booleanParam(name: 'KEYCLOAK',               defaultValue: false, description: 'Keycloak — IAM, SSO/OIDC/SAML [K8s]')
        booleanParam(name: 'DEX',                    defaultValue: false, description: 'Dex — OIDC federation and identity brokering [K8s]')
        booleanParam(name: 'AUTHENTIK',              defaultValue: false, description: 'Authentik — open source identity provider and SSO [K8s]')
        booleanParam(name: 'ZITADEL',                defaultValue: false, description: 'ZITADEL — cloud-native identity and access management [K8s]')
        booleanParam(name: 'AUTHELIA',               defaultValue: false, description: 'Authelia — authentication and authorization server, SSO/2FA [K8s]')
        booleanParam(name: 'SPIRE',                  defaultValue: false, description: 'SPIRE — SPIFFE workload identity runtime, zero-trust mTLS [K8s]')
        booleanParam(name: 'POMERIUM',               defaultValue: false, description: 'Pomerium — identity-aware proxy and zero-trust access gateway [K8s]')

        // ── Secrets Management (K8s) ──────────────────────────────────────
        booleanParam(name: 'VAULT',                  defaultValue: false, description: 'HashiCorp Vault — secrets management, dynamic credentials, PKI [K8s]')
        booleanParam(name: 'INFISICAL',              defaultValue: false, description: 'Infisical — open source secrets management platform [K8s]')

        // ── Policy Engines (K8s) ──────────────────────────────────────────
        booleanParam(name: 'OPA',                    defaultValue: false, description: 'OPA Gatekeeper — policy-as-code K8s admission controller [K8s]')
        booleanParam(name: 'KYVERNO',                defaultValue: false, description: 'Kyverno — Kubernetes-native policy engine (CNCF) [K8s]')
        booleanParam(name: 'KUBEWARDEN',             defaultValue: false, description: 'Kubewarden — WebAssembly-based K8s policy engine [K8s]')
        booleanParam(name: 'OPENFGA',                defaultValue: false, description: 'OpenFGA — fine-grained authorisation engine (Zanzibar model) [K8s]')

        // ── Runtime Security (K8s) ────────────────────────────────────────
        booleanParam(name: 'FALCO',                  defaultValue: false, description: 'Falco — eBPF runtime threat detection and alerting (CNCF) [K8s]')
        booleanParam(name: 'TETRAGON',               defaultValue: false, description: 'Tetragon — eBPF runtime security enforcement (Cilium/CNCF) [K8s]')
        booleanParam(name: 'TRACEE',                 defaultValue: false, description: 'Tracee — eBPF runtime security and forensics (Aqua Security) [K8s]')
        booleanParam(name: 'KUBEARMOR',              defaultValue: false, description: 'KubeArmor — runtime security enforcement using LSM + eBPF [K8s]')

        // ── WAF (K8s) ─────────────────────────────────────────────────────
        booleanParam(name: 'CORAZA_WAF',             defaultValue: false, description: 'Coraza WAF — open source OWASP-compatible Web Application Firewall [K8s]')

        // ── Certificates (K8s) ────────────────────────────────────────────
        booleanParam(name: 'CERT_MANAGER',           defaultValue: false, description: 'cert-manager — automated TLS certificate management (CNCF) [K8s]')

        // ── SAST Server (K8s) ─────────────────────────────────────────────
        booleanParam(name: 'SONARQUBE',              defaultValue: false, description: 'SonarQube Community — static code analysis and quality gates [K8s]')

        // ── Vulnerability Scanning (K8s) ──────────────────────────────────
        booleanParam(name: 'TRIVY_OPERATOR',         defaultValue: false, description: 'Trivy Operator — continuous in-cluster vulnerability scanning [K8s]')
        booleanParam(name: 'CLAIR',                  defaultValue: false, description: 'Clair — container image vulnerability static analysis [K8s]')
        booleanParam(name: 'OPENVAS',                defaultValue: false, description: 'OpenVAS/Greenbone — vulnerability scanner and management [K8s]')
        booleanParam(name: 'ANCHORE',                defaultValue: false, description: 'Anchore Engine — container vulnerability analysis server [K8s]')

        // ── DAST (K8s) ────────────────────────────────────────────────────
        booleanParam(name: 'ZAP',                    defaultValue: false, description: 'OWASP ZAP — dynamic application security testing server [K8s]')
        booleanParam(name: 'NUCLEI',                 defaultValue: false, description: 'Nuclei — fast template-based vulnerability scanner [K8s]')

        // ── Kubernetes Compliance (K8s) ───────────────────────────────────
        booleanParam(name: 'KUBESCAPE',              defaultValue: false, description: 'Kubescape — K8s risk and compliance scanning NSA/MITRE ATT&CK [K8s]')
        booleanParam(name: 'POLARIS',                defaultValue: false, description: 'Polaris — Kubernetes best practices validator and dashboard [K8s]')

        // ── Supply Chain (K8s) ────────────────────────────────────────────
        booleanParam(name: 'REKOR',                  defaultValue: false, description: 'Rekor — Sigstore immutable transparency log for supply chain [K8s]')
        booleanParam(name: 'FULCIO',                 defaultValue: false, description: 'Fulcio — Sigstore Certificate Authority for supply chain [K8s]')
        booleanParam(name: 'NOTARY',                 defaultValue: false, description: 'Notary/Notation — container image signing and content trust [K8s]')

        // ── Network Security (K8s) ────────────────────────────────────────
        booleanParam(name: 'SURICATA',               defaultValue: false, description: 'Suricata — high performance network IDS/IPS [K8s]')
        booleanParam(name: 'ZEEK',                   defaultValue: false, description: 'Zeek — open source network security monitor [K8s]')

        // ── SIEM / XDR (K8s) ─────────────────────────────────────────────
        booleanParam(name: 'WAZUH',                  defaultValue: false, description: 'Wazuh — open source XDR and SIEM platform [K8s]')

        // ── Vulnerability Management (K8s) ────────────────────────────────
        booleanParam(name: 'DEPENDENCY_TRACK',       defaultValue: false, description: 'Dependency Track — OWASP SCA and SBOM vulnerability management [K8s]')
        booleanParam(name: 'DEFECTDOJO',             defaultValue: false, description: 'DefectDojo — open source vulnerability management platform [K8s]')

        // ── SAST / Linters (CLI) ──────────────────────────────────────────
        booleanParam(name: 'PYLINT',                 defaultValue: false, description: 'Pylint — Python static code analysis [CLI]')
        booleanParam(name: 'PYFLAKES',               defaultValue: false, description: 'Pyflakes — lightweight Python source code checker [CLI]')
        booleanParam(name: 'FLAKE8',                 defaultValue: false, description: 'Flake8 — Python style guide enforcement PEP8 [CLI]')
        booleanParam(name: 'ESLINT',                 defaultValue: false, description: 'ESLint — JavaScript/TypeScript static analysis [CLI]')
        booleanParam(name: 'GOLANGCI',               defaultValue: false, description: 'GolangCI-Lint — Go linter aggregating 50+ linters [CLI]')
        booleanParam(name: 'SHELLCHECK',             defaultValue: false, description: 'ShellCheck — static analysis for shell scripts [CLI]')
        booleanParam(name: 'BANDIT',                 defaultValue: false, description: 'Bandit — Python security-focused static analysis [CLI]')
        booleanParam(name: 'BRAKEMAN',               defaultValue: false, description: 'Brakeman — Ruby on Rails static security analysis [CLI]')
        booleanParam(name: 'CODEQL',                 defaultValue: false, description: 'CodeQL — semantic code analysis engine (GitHub) [CLI]')
        booleanParam(name: 'SEMGREP',                defaultValue: false, description: 'Semgrep — multi-language pattern-based static analysis [CLI]')
        booleanParam(name: 'SPECTRAL',               defaultValue: false, description: 'Spectral — OpenAPI/AsyncAPI/JSON schema linting [CLI]')
        booleanParam(name: 'SNYK',                   defaultValue: false, description: 'Snyk — security scanning for code and dependencies [CLI]')

        // ── Dependency Scanning / SCA (CLI) ───────────────────────────────
        booleanParam(name: 'OWASP_DEP_CHECK',        defaultValue: false, description: 'OWASP Dependency-Check — known-vulnerable dependency scanner [CLI]')
        booleanParam(name: 'TRIVY_CLI',              defaultValue: false, description: 'Trivy CLI — filesystem and image vulnerability scanner [CLI]')
        booleanParam(name: 'GRYPE',                  defaultValue: false, description: 'Grype — container and filesystem vulnerability scanner [CLI]')
        booleanParam(name: 'SYFT',                   defaultValue: false, description: 'Syft — SBOM generation for container images and filesystems [CLI]')
        booleanParam(name: 'DOCKER_SCOUT',           defaultValue: false, description: 'Docker Scout — Docker-native image vulnerability scanning [CLI]')
        booleanParam(name: 'FOSSA',                  defaultValue: false, description: 'FOSSA — open source license compliance and security [CLI]')
        booleanParam(name: 'VULS',                   defaultValue: false, description: 'Vuls — agentless vulnerability scanner for Linux and containers [CLI]')
        booleanParam(name: 'OPENSCAP',               defaultValue: false, description: 'OpenSCAP — SCAP-based security compliance scanner [CLI]')

        // ── Image Signing (CLI) ───────────────────────────────────────────
        booleanParam(name: 'COSIGN',                 defaultValue: false, description: 'Cosign — container image signing and verification (Sigstore) [CLI]')

        // ── Secret Scanning (CLI) ─────────────────────────────────────────
        booleanParam(name: 'GITGUARDIAN',            defaultValue: false, description: 'GitGuardian — secrets detection in source code and git history [CLI]')
        booleanParam(name: 'GITLEAKS',               defaultValue: false, description: 'Gitleaks — secrets and sensitive data detection in git repos [CLI]')
        booleanParam(name: 'TRUFFLEHOG',             defaultValue: false, description: 'TruffleHog — leaked credentials and secrets in git history [CLI]')

        // ── IaC Security (CLI) ────────────────────────────────────────────
        booleanParam(name: 'KICS',                   defaultValue: false, description: 'KICS — infrastructure as code security scanning (Checkmarx) [CLI]')
        booleanParam(name: 'TERRASCAN',              defaultValue: false, description: 'Terrascan — Terraform IaC static security analysis [CLI]')
        booleanParam(name: 'TFSEC',                  defaultValue: false, description: 'tfsec — Terraform static analysis security scanner [CLI]')
        booleanParam(name: 'CHECKOV',                defaultValue: false, description: 'Checkov — IaC security scanning for Terraform, K8s, Helm, Docker [CLI]')

        // ── Kubernetes Security (CLI) ─────────────────────────────────────
        booleanParam(name: 'KUBE_BENCH',             defaultValue: false, description: 'kube-bench — CIS Kubernetes Benchmark compliance checker [CLI]')
        booleanParam(name: 'KUBE_HUNTER',            defaultValue: false, description: 'kube-hunter — Kubernetes penetration testing [CLI]')
        booleanParam(name: 'CNSPEC',                 defaultValue: false, description: 'cnspec — cloud and infrastructure security assessment (Mondoo) [CLI]')
        booleanParam(name: 'CNODE',                  defaultValue: false, description: 'cnode — Node.js dependency vulnerability audit [CLI]')
        booleanParam(name: 'KUBEAUDIT',              defaultValue: false, description: 'Kubeaudit — Kubernetes manifest and cluster security auditor [CLI]')

        // ── License Checking (CLI) ────────────────────────────────────────
        booleanParam(name: 'LICENSE_CHECK',          defaultValue: false, description: 'License Checker — open source license compliance scanning [CLI]')
        booleanParam(name: 'TERN',                   defaultValue: false, description: 'Tern — container image software composition and license analysis [CLI]')
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

        // ── INSTALL ───────────────────────────────────────────────────────

        stage('Install K8s Security Tools') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    // Identity & Access
                    if (params.KEYCLOAK)        { def s = load 'scripts/groovy/security-install-keycloak.groovy';        s() }
                    if (params.DEX)             { def s = load 'scripts/groovy/security-install-dex.groovy';             s() }
                    if (params.AUTHENTIK)       { def s = load 'scripts/groovy/security-install-authentik.groovy';       s() }
                    if (params.ZITADEL)         { def s = load 'scripts/groovy/security-install-zitadel.groovy';         s() }
                    if (params.AUTHELIA)        { def s = load 'scripts/groovy/security-install-authelia.groovy';        s() }
                    if (params.SPIRE)           { def s = load 'scripts/groovy/security-install-spire.groovy';           s() }
                    if (params.POMERIUM)        { def s = load 'scripts/groovy/security-install-pomerium.groovy';        s() }
                    // Secrets
                    if (params.VAULT)           { def s = load 'scripts/groovy/security-install-vault.groovy';           s() }
                    if (params.INFISICAL)       { def s = load 'scripts/groovy/security-install-infisical.groovy';       s() }
                    // Policy
                    if (params.OPA)             { def s = load 'scripts/groovy/security-install-opa.groovy';             s() }
                    if (params.KYVERNO)         { def s = load 'scripts/groovy/security-install-kyverno.groovy';         s() }
                    if (params.KUBEWARDEN)      { def s = load 'scripts/groovy/security-install-kubewarden.groovy';      s() }
                    if (params.OPENFGA)         { def s = load 'scripts/groovy/security-install-openfga.groovy';         s() }
                    // Runtime
                    if (params.FALCO)           { def s = load 'scripts/groovy/security-install-falco.groovy';           s() }
                    if (params.TETRAGON)        { def s = load 'scripts/groovy/security-install-tetragon.groovy';        s() }
                    if (params.TRACEE)          { def s = load 'scripts/groovy/security-install-tracee.groovy';          s() }
                    if (params.KUBEARMOR)       { def s = load 'scripts/groovy/security-install-kubearmor.groovy';       s() }
                    // WAF
                    if (params.CORAZA_WAF)      { def s = load 'scripts/groovy/security-install-coraza-waf.groovy';      s() }
                    // Certificates
                    if (params.CERT_MANAGER)    { def s = load 'scripts/groovy/security-install-cert-manager.groovy';    s() }
                    // SAST server
                    if (params.SONARQUBE)       { def s = load 'scripts/groovy/security-install-sonarqube.groovy';       s() }
                    // Vulnerability Scanning
                    if (params.TRIVY_OPERATOR)  { def s = load 'scripts/groovy/security-install-trivy-operator.groovy';  s() }
                    if (params.CLAIR)           { def s = load 'scripts/groovy/security-install-clair.groovy';           s() }
                    if (params.OPENVAS)         { def s = load 'scripts/groovy/security-install-openvas.groovy';         s() }
                    if (params.ANCHORE)         { def s = load 'scripts/groovy/security-install-anchore.groovy';         s() }
                    // DAST
                    if (params.ZAP)             { def s = load 'scripts/groovy/security-install-zap.groovy';             s() }
                    if (params.NUCLEI)          { def s = load 'scripts/groovy/security-install-nuclei.groovy';          s() }
                    // K8s Compliance
                    if (params.KUBESCAPE)       { def s = load 'scripts/groovy/security-install-kubescape.groovy';       s() }
                    if (params.POLARIS)         { def s = load 'scripts/groovy/security-install-polaris.groovy';         s() }
                    // Supply Chain
                    if (params.REKOR)           { def s = load 'scripts/groovy/security-install-rekor.groovy';           s() }
                    if (params.FULCIO)          { def s = load 'scripts/groovy/security-install-fulcio.groovy';          s() }
                    if (params.NOTARY)          { def s = load 'scripts/groovy/security-install-notary.groovy';          s() }
                    // Network Security
                    if (params.SURICATA)        { def s = load 'scripts/groovy/security-install-suricata.groovy';        s() }
                    if (params.ZEEK)            { def s = load 'scripts/groovy/security-install-zeek.groovy';            s() }
                    // SIEM
                    if (params.WAZUH)           { def s = load 'scripts/groovy/security-install-wazuh.groovy';           s() }
                    // Vulnerability Management
                    if (params.DEPENDENCY_TRACK){ def s = load 'scripts/groovy/security-install-dependency-track.groovy';s() }
                    if (params.DEFECTDOJO)      { def s = load 'scripts/groovy/security-install-defectdojo.groovy';      s() }
                }
            }
        }

        stage('Pull CLI Security Tool Images') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    // SAST / Linters
                    if (params.PYLINT || params.PYFLAKES || params.FLAKE8 || params.BANDIT) {
                        sh 'docker pull python:3.13-slim'
                    }
                    if (params.ESLINT || params.LICENSE_CHECK || params.CNODE) {
                        sh 'docker pull node:22-alpine'
                    }
                    if (params.GOLANGCI)     { sh 'docker pull golangci/golangci-lint:latest' }
                    if (params.SHELLCHECK)   { sh 'docker pull koalaman/shellcheck:stable' }
                    if (params.BRAKEMAN)     { sh 'docker pull presidentbeef/brakeman:latest' }
                    if (params.CODEQL)       { sh 'docker pull ghcr.io/github/codeql-action:latest || true' }
                    if (params.SEMGREP)      { sh 'docker pull semgrep/semgrep:latest' }
                    if (params.SPECTRAL)     { sh 'docker pull stoplight/spectral:latest' }
                    if (params.SNYK)         { sh 'docker pull snyk/snyk:latest' }
                    // Dependency / SCA
                    if (params.OWASP_DEP_CHECK) { sh 'docker pull owasp/dependency-check:latest' }
                    if (params.TRIVY_CLI)    { sh 'docker pull aquasec/trivy:latest' }
                    if (params.GRYPE)        { sh 'docker pull anchore/grype:latest' }
                    if (params.SYFT)         { sh 'docker pull anchore/syft:latest' }
                    if (params.DOCKER_SCOUT) { sh 'docker pull docker/scout-cli:latest' }
                    if (params.FOSSA)        { sh 'docker pull fossas/fossa-cli:latest' }
                    if (params.VULS)         { sh 'docker pull vuls/vuls:latest' }
                    if (params.OPENSCAP)     { sh 'docker pull openscap/openscap:latest' }
                    // Image Signing
                    if (params.COSIGN)       { sh 'docker pull gcr.io/projectsigstore/cosign:latest' }
                    // Secret Scanning
                    if (params.GITGUARDIAN)  { sh 'docker pull gitguardian/ggshield:latest' }
                    if (params.GITLEAKS)     { sh 'docker pull zricethezav/gitleaks:latest' }
                    if (params.TRUFFLEHOG)   { sh 'docker pull trufflesecurity/trufflehog:latest' }
                    // IaC
                    if (params.KICS)         { sh 'docker pull checkmarx/kics:latest' }
                    if (params.TERRASCAN)    { sh 'docker pull tenable/terrascan:latest' }
                    if (params.TFSEC)        { sh 'docker pull aquasec/tfsec:latest' }
                    if (params.CHECKOV)      { sh 'docker pull bridgecrew/checkov:latest' }
                    // K8s Security
                    if (params.KUBE_BENCH)   { sh 'docker pull aquasec/kube-bench:latest' }
                    if (params.KUBE_HUNTER)  { sh 'docker pull aquasec/kube-hunter:latest' }
                    if (params.CNSPEC)       { sh 'docker pull mondoo/cnspec:latest' }
                    if (params.KUBEAUDIT)    { sh 'docker pull shopify/kubeaudit:latest' }
                    // License
                    if (params.TERN)         { sh 'docker pull philipssoftware/tern:latest' }
                }
            }
        }

        // ── UNINSTALL ─────────────────────────────────────────────────────

        stage('Uninstall K8s Security Tools') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.KEYCLOAK)        { sh 'helm uninstall keycloak         -n keycloak         --ignore-not-found || true' }
                    if (params.DEX)             { sh 'helm uninstall dex              -n dex              --ignore-not-found || true' }
                    if (params.AUTHENTIK)       { sh 'helm uninstall authentik        -n authentik        --ignore-not-found || true' }
                    if (params.ZITADEL)         { sh 'helm uninstall zitadel          -n zitadel          --ignore-not-found || true' }
                    if (params.AUTHELIA)        { sh 'helm uninstall authelia         -n authelia         --ignore-not-found || true' }
                    if (params.SPIRE)           { sh 'helm uninstall spire            -n spire            --ignore-not-found || true' }
                    if (params.POMERIUM)        { sh 'helm uninstall pomerium         -n pomerium         --ignore-not-found || true' }
                    if (params.VAULT)           { sh 'helm uninstall vault            -n vault            --ignore-not-found || true' }
                    if (params.INFISICAL)       { sh 'helm uninstall infisical        -n infisical        --ignore-not-found || true' }
                    if (params.OPA)             { sh 'helm uninstall opa              -n opa              --ignore-not-found || true' }
                    if (params.KYVERNO)         { sh 'helm uninstall kyverno          -n kyverno          --ignore-not-found || true' }
                    if (params.KUBEWARDEN)      { sh 'helm uninstall kubewarden       -n kubewarden       --ignore-not-found || true' }
                    if (params.OPENFGA)         { sh 'helm uninstall openfga          -n openfga          --ignore-not-found || true' }
                    if (params.FALCO)           { sh 'helm uninstall falco            -n falco            --ignore-not-found || true' }
                    if (params.TETRAGON)        { sh 'helm uninstall tetragon         -n tetragon         --ignore-not-found || true' }
                    if (params.TRACEE)          { sh 'helm uninstall tracee           -n tracee           --ignore-not-found || true' }
                    if (params.KUBEARMOR)       { sh 'helm uninstall kubearmor        -n kubearmor        --ignore-not-found || true' }
                    if (params.CORAZA_WAF)      { sh 'helm uninstall coraza-waf       -n coraza-waf       --ignore-not-found || true' }
                    if (params.CERT_MANAGER)    { sh 'helm uninstall cert-manager     -n cert-manager     --ignore-not-found || true' }
                    if (params.SONARQUBE)       { sh 'helm uninstall sonarqube        -n sonarqube        --ignore-not-found || true' }
                    if (params.TRIVY_OPERATOR)  { sh 'helm uninstall trivy-operator   -n trivy-operator   --ignore-not-found || true' }
                    if (params.CLAIR)           { sh 'helm uninstall clair            -n clair            --ignore-not-found || true' }
                    if (params.OPENVAS)         { sh 'helm uninstall openvas          -n openvas          --ignore-not-found || true' }
                    if (params.ANCHORE)         { sh 'helm uninstall anchore          -n anchore          --ignore-not-found || true' }
                    if (params.ZAP)             { sh 'helm uninstall zap              -n zap              --ignore-not-found || true' }
                    if (params.NUCLEI)          { sh 'helm uninstall nuclei           -n nuclei           --ignore-not-found || true' }
                    if (params.KUBESCAPE)       { sh 'helm uninstall kubescape        -n kubescape        --ignore-not-found || true' }
                    if (params.POLARIS)         { sh 'helm uninstall polaris          -n polaris          --ignore-not-found || true' }
                    if (params.REKOR)           { sh 'helm uninstall rekor            -n rekor            --ignore-not-found || true' }
                    if (params.FULCIO)          { sh 'helm uninstall fulcio           -n fulcio           --ignore-not-found || true' }
                    if (params.NOTARY)          { sh 'helm uninstall notary           -n notary           --ignore-not-found || true' }
                    if (params.SURICATA)        { sh 'helm uninstall suricata         -n suricata         --ignore-not-found || true' }
                    if (params.ZEEK)            { sh 'helm uninstall zeek             -n zeek             --ignore-not-found || true' }
                    if (params.WAZUH)           { sh 'helm uninstall wazuh            -n wazuh            --ignore-not-found || true' }
                    if (params.DEPENDENCY_TRACK){ sh 'helm uninstall dependency-track -n dependency-track --ignore-not-found || true' }
                    if (params.DEFECTDOJO)      { sh 'helm uninstall defectdojo       -n defectdojo       --ignore-not-found || true' }
                }
            }
        }

        stage('Remove CLI Tool Images') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.PYLINT || params.PYFLAKES || params.FLAKE8 || params.BANDIT) {
                        sh 'docker rmi python:3.13-slim || true'
                    }
                    if (params.ESLINT || params.LICENSE_CHECK || params.CNODE) {
                        sh 'docker rmi node:22-alpine || true'
                    }
                    if (params.GOLANGCI)     { sh 'docker rmi golangci/golangci-lint:latest || true' }
                    if (params.SHELLCHECK)   { sh 'docker rmi koalaman/shellcheck:stable || true' }
                    if (params.BRAKEMAN)     { sh 'docker rmi presidentbeef/brakeman:latest || true' }
                    if (params.CODEQL)       { sh 'docker rmi ghcr.io/github/codeql-action:latest || true' }
                    if (params.SEMGREP)      { sh 'docker rmi semgrep/semgrep:latest || true' }
                    if (params.SPECTRAL)     { sh 'docker rmi stoplight/spectral:latest || true' }
                    if (params.SNYK)         { sh 'docker rmi snyk/snyk:latest || true' }
                    if (params.OWASP_DEP_CHECK) { sh 'docker rmi owasp/dependency-check:latest || true' }
                    if (params.TRIVY_CLI)    { sh 'docker rmi aquasec/trivy:latest || true' }
                    if (params.GRYPE)        { sh 'docker rmi anchore/grype:latest || true' }
                    if (params.SYFT)         { sh 'docker rmi anchore/syft:latest || true' }
                    if (params.DOCKER_SCOUT) { sh 'docker rmi docker/scout-cli:latest || true' }
                    if (params.FOSSA)        { sh 'docker rmi fossas/fossa-cli:latest || true' }
                    if (params.VULS)         { sh 'docker rmi vuls/vuls:latest || true' }
                    if (params.OPENSCAP)     { sh 'docker rmi openscap/openscap:latest || true' }
                    if (params.COSIGN)       { sh 'docker rmi gcr.io/projectsigstore/cosign:latest || true' }
                    if (params.GITGUARDIAN)  { sh 'docker rmi gitguardian/ggshield:latest || true' }
                    if (params.GITLEAKS)     { sh 'docker rmi zricethezav/gitleaks:latest || true' }
                    if (params.TRUFFLEHOG)   { sh 'docker rmi trufflesecurity/trufflehog:latest || true' }
                    if (params.KICS)         { sh 'docker rmi checkmarx/kics:latest || true' }
                    if (params.TERRASCAN)    { sh 'docker rmi tenable/terrascan:latest || true' }
                    if (params.TFSEC)        { sh 'docker rmi aquasec/tfsec:latest || true' }
                    if (params.CHECKOV)      { sh 'docker rmi bridgecrew/checkov:latest || true' }
                    if (params.KUBE_BENCH)   { sh 'docker rmi aquasec/kube-bench:latest || true' }
                    if (params.KUBE_HUNTER)  { sh 'docker rmi aquasec/kube-hunter:latest || true' }
                    if (params.CNSPEC)       { sh 'docker rmi mondoo/cnspec:latest || true' }
                    if (params.KUBEAUDIT)    { sh 'docker rmi shopify/kubeaudit:latest || true' }
                    if (params.TERN)         { sh 'docker rmi philipssoftware/tern:latest || true' }
                }
            }
        }

        // ── CONFIGURE ─────────────────────────────────────────────────────

        stage('Configure Security Tools') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps {
                script {
                    // Identity & Access
                    if (params.KEYCLOAK)        { def s = load 'scripts/groovy/security-configure-keycloak.groovy';        s() }
                    if (params.AUTHENTIK)       { def s = load 'scripts/groovy/security-configure-authentik.groovy';       s() }
                    if (params.ZITADEL)         { def s = load 'scripts/groovy/security-configure-zitadel.groovy';         s() }
                    if (params.AUTHELIA)        { def s = load 'scripts/groovy/security-configure-authelia.groovy';        s() }
                    if (params.POMERIUM)        { def s = load 'scripts/groovy/security-configure-pomerium.groovy';        s() }
                    // Secrets
                    if (params.VAULT)           { def s = load 'scripts/groovy/security-configure-vault.groovy';           s() }
                    if (params.INFISICAL)       { def s = load 'scripts/groovy/security-configure-infisical.groovy';       s() }
                    // SAST server
                    if (params.SONARQUBE)       { def s = load 'scripts/groovy/security-configure-sonarqube.groovy';       s() }
                    // Vulnerability Scanning
                    if (params.ANCHORE)         { def s = load 'scripts/groovy/security-configure-anchore.groovy';         s() }
                    if (params.OPENVAS)         { def s = load 'scripts/groovy/security-configure-openvas.groovy';         s() }
                    // DAST
                    if (params.ZAP)             { def s = load 'scripts/groovy/security-configure-zap.groovy';             s() }
                    if (params.NUCLEI)          { def s = load 'scripts/groovy/security-configure-nuclei.groovy';          s() }
                    // SIEM
                    if (params.WAZUH)           { def s = load 'scripts/groovy/security-configure-wazuh.groovy';           s() }
                    // Vulnerability Management
                    if (params.DEPENDENCY_TRACK){ def s = load 'scripts/groovy/security-configure-dependency-track.groovy';s() }
                    if (params.DEFECTDOJO)      { def s = load 'scripts/groovy/security-configure-defectdojo.groovy';      s() }
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
            sh 'docker image prune -f 2>/dev/null || true'
        }
    }
}
