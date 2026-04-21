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
            choices: ['INSTALL_K8S_REPOS', 'UNINSTALL_K8S_REPOS', 'CONFIGURE_K8S_REPOS', 'CREATE_CLOUD_REGISTRY', 'DESTROY_CLOUD_REGISTRY'],
            description: 'INSTALL_K8S_REPOS/UNINSTALL_K8S_REPOS — deploy or remove selected repo tools on K8s. CONFIGURE_K8S_REPOS — post-install setup (projects, replication, webhooks). CREATE/DESTROY — manage cloud-native container registry (ECR/GCR/ACR).'
        )

        // Container Registries
        booleanParam(name: 'HARBOR',       defaultValue: false, description: 'Harbor — container registry with scanning, RBAC, replication')
        booleanParam(name: 'ZOT',          defaultValue: false, description: 'Zot — lightweight OCI-native registry (CNCF)')
        booleanParam(name: 'DISTRIBUTION', defaultValue: false, description: 'Distribution — minimal Docker/OCI registry')
        booleanParam(name: 'QUAY',         defaultValue: false, description: 'Quay — enterprise container registry by Red Hat')
        booleanParam(name: 'KRAKEN',       defaultValue: false, description: 'Kraken — P2P container registry for large-scale distribution')
        booleanParam(name: 'DRAGONFLY',    defaultValue: false, description: 'Dragonfly — CNCF P2P image distribution accelerator')

        // Universal Artifact Repositories
        booleanParam(name: 'NEXUS',        defaultValue: false, description: 'Nexus OSS — Maven, npm, PyPI, Go, Docker, Helm, Cargo, NuGet')
        booleanParam(name: 'PULP',         defaultValue: false, description: 'Pulp — Docker, RPM, Deb, Python, Ansible, File')

        // Git + Package Registries
        booleanParam(name: 'GITEA',        defaultValue: false, description: 'Gitea — Git server + npm, Maven, PyPI, Go, Helm, Docker, Cargo')
        booleanParam(name: 'FORGEJO',      defaultValue: false, description: 'Forgejo — Gitea fork with community governance')
        booleanParam(name: 'GOGS',         defaultValue: false, description: 'Gogs — lightweight self-hosted Git service')
        booleanParam(name: 'GITBUCKET',    defaultValue: false, description: 'GitBucket — GitHub-like Git platform on JVM')
        booleanParam(name: 'ONEDEV',       defaultValue: false, description: 'OneDev — Git + CI + package registry in one')
        booleanParam(name: 'GITLAB',       defaultValue: false, description: 'GitLab CE — Git, CI, container registry, package registry')

        // Helm Chart Repositories
        booleanParam(name: 'CHARTMUSEUM',  defaultValue: false, description: 'ChartMuseum — dedicated Helm chart repository server')
        booleanParam(name: 'TERRAREG',     defaultValue: false, description: 'Terrareg — Terraform module registry')

        // npm / Node.js
        booleanParam(name: 'VERDACCIO',    defaultValue: false, description: 'Verdaccio — private npm/yarn registry')
        booleanParam(name: 'CNPMJS',       defaultValue: false, description: 'Cnpmjs — npm mirror registry server')

        // Python
        booleanParam(name: 'PYPISERVER',   defaultValue: false, description: 'Pypiserver — minimal PyPI-compatible package server')
        booleanParam(name: 'DEVPI',        defaultValue: false, description: 'Devpi — PyPI server with caching and staging')
        booleanParam(name: 'QUETZ',        defaultValue: false, description: 'Quetz — open source Conda package server')

        // Go
        booleanParam(name: 'ATHENS',       defaultValue: false, description: 'Athens — Go module proxy and registry')
        booleanParam(name: 'GOPROXY',      defaultValue: false, description: 'Goproxy — simple Go module proxy server')

        // Java / JVM
        booleanParam(name: 'REPOSILITE',   defaultValue: false, description: 'Reposilite — lightweight Maven/Gradle repository')

        // .NET
        booleanParam(name: 'BAGET',        defaultValue: false, description: 'BaGet — lightweight NuGet package server')

        // Rust
        booleanParam(name: 'KELLNR',       defaultValue: false, description: 'Kellnr — private Cargo/Rust crate registry')
        booleanParam(name: 'ALEXANDRIE',   defaultValue: false, description: 'Alexandrie — alternative Cargo/Rust crate registry')

        // Ruby
        booleanParam(name: 'GEMINABOX',    defaultValue: false, description: 'Geminabox — private Ruby gem server')

        // C/C++
        booleanParam(name: 'CONAN_SERVER', defaultValue: false, description: 'Conan Server — C/C++ package repository')

        // Debian/APT
        booleanParam(name: 'APTLY',        defaultValue: false, description: 'Aptly — Debian/APT package repository manager')
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
                    echo "Creating cloud registry for: ${cloud}"
                    if (cloud == 'AWS') {
                        def s = load 'scripts/groovy/repo-create-ecr.groovy';             s()
                    } else if (cloud == 'GCP') {
                        def s = load 'scripts/groovy/repo-create-artifact-registry.groovy'; s()
                    } else if (cloud == 'AZURE') {
                        def s = load 'scripts/groovy/repo-create-acr.groovy';             s()
                    } else {
                        error "Unsupported cloud provider: ${cloud}. Auto-detection failed — check cloud metadata service."
                    }
                }
            }
        }

        stage('Destroy Cloud Registry') {
            when { expression { params.ACTION == 'DESTROY_CLOUD_REGISTRY' } }
            steps {
                script {
                    def cloud = env.CLOUD_PROVIDER
                    echo "Destroying cloud registry for: ${cloud}"
                    if (cloud == 'AWS') {
                        def s = load 'scripts/groovy/repo-delete-ecr.groovy';             s()
                    } else if (cloud == 'GCP') {
                        def s = load 'scripts/groovy/repo-delete-artifact-registry.groovy'; s()
                    } else if (cloud == 'AZURE') {
                        def s = load 'scripts/groovy/repo-delete-acr.groovy';             s()
                    } else {
                        error "Unsupported cloud provider: ${cloud}. Auto-detection failed — check cloud metadata service."
                    }
                }
            }
        }

        stage('Load Kubeconfig') {
            when { expression { params.ACTION == 'INSTALL_K8S_REPOS' } }
            steps {
                script {
                    def kubeconfigContent = ''
                    if (fileExists('infra.env')) {
                        kubeconfigContent = readFile('infra.env').trim()
                            .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last() ?: ''
                    }
                    if (!kubeconfigContent) {
                        error "KUBECONFIG_CONTENT not found in infra.env — run the k8s pipeline first to provision a cluster"
                    }
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kubeconfigContent
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig"
                    sh "rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        stage('Install Selected Repos') {
            when { expression { params.ACTION == 'INSTALL_K8S_REPOS' } }
            steps {
                script {
                    if (params.HARBOR)       { def s = load 'scripts/groovy/install-harbor.groovy';       s() }
                    if (params.ZOT)          { def s = load 'scripts/groovy/install-zot.groovy';          s() }
                    if (params.DISTRIBUTION) { def s = load 'scripts/groovy/install-distribution.groovy'; s() }
                    if (params.QUAY)         { def s = load 'scripts/groovy/install-quay.groovy';         s() }
                    if (params.KRAKEN)       { def s = load 'scripts/groovy/install-kraken.groovy';       s() }
                    if (params.DRAGONFLY)    { def s = load 'scripts/groovy/install-dragonfly.groovy';    s() }
                    if (params.NEXUS)        { def s = load 'scripts/groovy/install-nexus.groovy';        s() }
                    if (params.PULP)         { def s = load 'scripts/groovy/install-pulp.groovy';         s() }
                    if (params.GITEA)        { def s = load 'scripts/groovy/install-gitea.groovy';        s() }
                    if (params.FORGEJO)      { def s = load 'scripts/groovy/install-forgejo.groovy';      s() }
                    if (params.GOGS)         { def s = load 'scripts/groovy/install-gogs.groovy';         s() }
                    if (params.GITBUCKET)    { def s = load 'scripts/groovy/install-gitbucket.groovy';    s() }
                    if (params.ONEDEV)       { def s = load 'scripts/groovy/install-onedev.groovy';       s() }
                    if (params.GITLAB)       { def s = load 'scripts/groovy/install-gitlab.groovy';       s() }
                    if (params.CHARTMUSEUM)  { def s = load 'scripts/groovy/install-chartmuseum.groovy';  s() }
                    if (params.TERRAREG)     { def s = load 'scripts/groovy/install-terrareg.groovy';     s() }
                    if (params.VERDACCIO)    { def s = load 'scripts/groovy/install-verdaccio.groovy';    s() }
                    if (params.CNPMJS)       { def s = load 'scripts/groovy/install-cnpmjs.groovy';       s() }
                    if (params.PYPISERVER)   { def s = load 'scripts/groovy/install-pypiserver.groovy';   s() }
                    if (params.DEVPI)        { def s = load 'scripts/groovy/install-devpi.groovy';        s() }
                    if (params.QUETZ)        { def s = load 'scripts/groovy/install-quetz.groovy';        s() }
                    if (params.ATHENS)       { def s = load 'scripts/groovy/install-athens.groovy';       s() }
                    if (params.GOPROXY)      { def s = load 'scripts/groovy/install-goproxy.groovy';      s() }
                    if (params.REPOSILITE)   { def s = load 'scripts/groovy/install-reposilite.groovy';   s() }
                    if (params.BAGET)        { def s = load 'scripts/groovy/install-baget.groovy';        s() }
                    if (params.KELLNR)       { def s = load 'scripts/groovy/install-kellnr.groovy';       s() }
                    if (params.ALEXANDRIE)   { def s = load 'scripts/groovy/install-alexandrie.groovy';   s() }
                    if (params.GEMINABOX)    { def s = load 'scripts/groovy/install-geminabox.groovy';    s() }
                    if (params.CONAN_SERVER) { def s = load 'scripts/groovy/install-conan-server.groovy'; s() }
                    if (params.APTLY)        { def s = load 'scripts/groovy/install-aptly.groovy';        s() }
                }
            }
        }

        stage('Uninstall Selected Repos') {
            when { expression { params.ACTION == 'UNINSTALL_K8S_REPOS' } }
            steps {
                script {
                    if (params.HARBOR)       { sh 'helm uninstall harbor        -n harbor        --ignore-not-found || true' }
                    if (params.ZOT)          { sh 'helm uninstall zot           -n zot           --ignore-not-found || true' }
                    if (params.DISTRIBUTION) { sh 'helm uninstall distribution  -n distribution  --ignore-not-found || true' }
                    if (params.QUAY)         { sh 'helm uninstall quay          -n quay          --ignore-not-found || true' }
                    if (params.KRAKEN)       { sh 'helm uninstall kraken        -n kraken        --ignore-not-found || true' }
                    if (params.DRAGONFLY)    { sh 'helm uninstall dragonfly     -n dragonfly     --ignore-not-found || true' }
                    if (params.NEXUS)        { sh 'helm uninstall nexus         -n nexus         --ignore-not-found || true' }
                    if (params.PULP)         { sh 'helm uninstall pulp          -n pulp          --ignore-not-found || true' }
                    if (params.GITEA)        { sh 'helm uninstall gitea         -n gitea         --ignore-not-found || true' }
                    if (params.FORGEJO)      { sh 'helm uninstall forgejo       -n forgejo       --ignore-not-found || true' }
                    if (params.GOGS)         { sh 'helm uninstall gogs          -n gogs          --ignore-not-found || true' }
                    if (params.GITBUCKET)    { sh 'helm uninstall gitbucket     -n gitbucket     --ignore-not-found || true' }
                    if (params.ONEDEV)       { sh 'helm uninstall onedev        -n onedev        --ignore-not-found || true' }
                    if (params.GITLAB)       { sh 'helm uninstall gitlab        -n gitlab        --ignore-not-found || true' }
                    if (params.CHARTMUSEUM)  { sh 'helm uninstall chartmuseum   -n chartmuseum   --ignore-not-found || true' }
                    if (params.TERRAREG)     { sh 'helm uninstall terrareg      -n terrareg      --ignore-not-found || true' }
                    if (params.VERDACCIO)    { sh 'helm uninstall verdaccio     -n verdaccio     --ignore-not-found || true' }
                    if (params.CNPMJS)       { sh 'helm uninstall cnpmjs        -n cnpmjs        --ignore-not-found || true' }
                    if (params.PYPISERVER)   { sh 'helm uninstall pypiserver    -n pypiserver    --ignore-not-found || true' }
                    if (params.DEVPI)        { sh 'helm uninstall devpi         -n devpi         --ignore-not-found || true' }
                    if (params.QUETZ)        { sh 'helm uninstall quetz         -n quetz         --ignore-not-found || true' }
                    if (params.ATHENS)       { sh 'helm uninstall athens        -n athens        --ignore-not-found || true' }
                    if (params.GOPROXY)      { sh 'helm uninstall goproxy       -n goproxy       --ignore-not-found || true' }
                    if (params.REPOSILITE)   { sh 'helm uninstall reposilite    -n reposilite    --ignore-not-found || true' }
                    if (params.BAGET)        { sh 'helm uninstall baget         -n baget         --ignore-not-found || true' }
                    if (params.KELLNR)       { sh 'helm uninstall kellnr        -n kellnr        --ignore-not-found || true' }
                    if (params.ALEXANDRIE)   { sh 'helm uninstall alexandrie    -n alexandrie    --ignore-not-found || true' }
                    if (params.GEMINABOX)    { sh 'helm uninstall geminabox     -n geminabox     --ignore-not-found || true' }
                    if (params.CONAN_SERVER) { sh 'helm uninstall conan-server  -n conan-server  --ignore-not-found || true' }
                    if (params.APTLY)        { sh 'helm uninstall aptly         -n aptly         --ignore-not-found || true' }
                }
            }
        }

        stage('Configure Selected Repos') {
            when { expression { params.ACTION == 'CONFIGURE_K8S_REPOS' } }
            steps {
                script {
                    // Container registries — projects, replication rules, webhooks, robot accounts
                    if (params.HARBOR)       { def s = load 'scripts/groovy/registry-configure-harbor.groovy';      s() }
                    if (params.ZOT)          { def s = load 'scripts/groovy/registry-configure-zot.groovy';         s() }
                    // Universal artifact repos — proxy repos, hosted repos, blob stores
                    if (params.NEXUS)        { def s = load 'scripts/groovy/registry-configure-nexus.groovy';       s() }
                    // Git servers — organisations, repos, webhooks, SSH keys
                    if (params.GITEA)        { def s = load 'scripts/groovy/registry-configure-gitea.groovy';       s() }
                    if (params.GITLAB)       { def s = load 'scripts/groovy/registry-configure-gitlab.groovy';      s() }
                    if (params.FORGEJO)      { def s = load 'scripts/groovy/registry-configure-forgejo.groovy';     s() }
                    // Helm chart repo
                    if (params.CHARTMUSEUM)  { def s = load 'scripts/groovy/registry-configure-chartmuseum.groovy'; s() }
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
