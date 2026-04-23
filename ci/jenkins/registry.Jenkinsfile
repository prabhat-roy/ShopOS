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
        booleanParam(name: 'HARBOR',        defaultValue: false,  description: 'Harbor — enterprise container registry')
        booleanParam(name: 'ZOT',           defaultValue: false,  description: 'Zot — OCI-native lightweight registry')
        booleanParam(name: 'DISTRIBUTION',  defaultValue: false,  description: 'Distribution — CNCF reference OCI registry')
        booleanParam(name: 'QUAY',          defaultValue: false,  description: 'Quay — Red Hat enterprise registry')
        booleanParam(name: 'KRAKEN',        defaultValue: false,  description: 'Kraken — P2P Docker registry')
        booleanParam(name: 'DRAGONFLY',     defaultValue: false,  description: 'Dragonfly — P2P image distribution')
        booleanParam(name: 'NEXUS',         defaultValue: false,  description: 'Nexus — universal artifact repository')
        booleanParam(name: 'PULP',          defaultValue: false,  description: 'Pulp — on-prem package management')
        booleanParam(name: 'GITEA',         defaultValue: false,  description: 'Gitea — lightweight self-hosted Git')
        booleanParam(name: 'FORGEJO',       defaultValue: false,  description: 'Forgejo — community-driven Gitea fork')
        booleanParam(name: 'GOGS',          defaultValue: false,  description: 'Gogs — minimal self-hosted Git')
        booleanParam(name: 'GITBUCKET',     defaultValue: false,  description: 'GitBucket — GitHub-compatible Git platform')
        booleanParam(name: 'ONEDEV',        defaultValue: false,  description: 'OneDev — Git server with built-in CI/CD')
        booleanParam(name: 'GITLAB',        defaultValue: false, description: 'GitLab — full DevOps platform (heavy)')
        booleanParam(name: 'CHARTMUSEUM',   defaultValue: false,  description: 'ChartMuseum — Helm chart repository')
        booleanParam(name: 'TERRAREG',      defaultValue: false,  description: 'Terrareg — Terraform module registry')
        booleanParam(name: 'VERDACCIO',     defaultValue: false,  description: 'Verdaccio — npm/yarn private registry')
        booleanParam(name: 'CNPMJS',        defaultValue: false,  description: 'Cnpmjs — npm private registry')
        booleanParam(name: 'PYPISERVER',    defaultValue: false,  description: 'Pypiserver — minimal PyPI server')
        booleanParam(name: 'DEVPI',         defaultValue: false,  description: 'Devpi — full-featured PyPI server')
        booleanParam(name: 'QUETZ',         defaultValue: false,  description: 'Quetz — conda package server')
        booleanParam(name: 'ATHENS',        defaultValue: false,  description: 'Athens — Go module proxy')
        booleanParam(name: 'GOPROXY',       defaultValue: false,  description: 'Goproxy — simple Go module proxy')
        booleanParam(name: 'REPOSILITE',    defaultValue: false,  description: 'Reposilite — lightweight Maven/Gradle repo')
        booleanParam(name: 'BAGET',         defaultValue: false,  description: 'BaGet — NuGet-compatible package server')
        booleanParam(name: 'KELLNR',        defaultValue: false,  description: 'Kellnr — Rust crate registry')
        booleanParam(name: 'ALEXANDRIE',    defaultValue: false,  description: 'Alexandrie — Rust crate registry alternative')
        booleanParam(name: 'GEMINABOX',     defaultValue: false,  description: 'Geminabox — Ruby gem server')
        booleanParam(name: 'CONAN_SERVER',  defaultValue: false,  description: 'Conan Server — C/C++ package registry')
        booleanParam(name: 'APTLY',         defaultValue: false,  description: 'Aptly — Debian/Ubuntu apt repository')
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
                    if (cloud == 'AWS')        { def s = load 'scripts/groovy/repo-create-ecr.groovy'; s() }
                    else if (cloud == 'GCP')   { def s = load 'scripts/groovy/repo-create-artifact-registry.groovy'; s() }
                    else if (cloud == 'AZURE') { def s = load 'scripts/groovy/repo-create-acr.groovy'; s() }
                    else                       { error "Unsupported cloud provider: ${cloud}" }
                }
            }
        }

        stage('Destroy Cloud Registry') {
            when { expression { params.ACTION == 'DESTROY_CLOUD_REGISTRY' } }
            steps {
                script {
                    def cloud = env.CLOUD_PROVIDER
                    if (cloud == 'AWS')        { def s = load 'scripts/groovy/repo-delete-ecr.groovy'; s() }
                    else if (cloud == 'GCP')   { def s = load 'scripts/groovy/repo-delete-artifact-registry.groovy'; s() }
                    else if (cloud == 'AZURE') { def s = load 'scripts/groovy/repo-delete-acr.groovy'; s() }
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

        stage('Container Registries') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'
                    if (params.HARBOR)       { def s = load 'scripts/groovy/install-harbor.groovy'; s(); def c = load 'scripts/groovy/registry-configure-harbor.groovy'; c(); e('harbor') }
                    if (params.ZOT)          { def s = load 'scripts/groovy/install-zot.groovy'; s(); def c = load 'scripts/groovy/registry-configure-zot.groovy'; c(); e('zot') }
                    if (params.DISTRIBUTION) { def s = load 'scripts/groovy/install-distribution.groovy'; s(); e('distribution') }
                    if (params.QUAY)         { def s = load 'scripts/groovy/install-quay.groovy'; s(); e('quay') }
                    if (params.KRAKEN)       { def s = load 'scripts/groovy/install-kraken.groovy'; s(); e('kraken') }
                    if (params.DRAGONFLY)    { def s = load 'scripts/groovy/install-dragonfly.groovy'; s(); e('dragonfly') }
                }
            }
        }

        stage('Artifact Repositories') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'
                    if (params.NEXUS) { def s = load 'scripts/groovy/install-nexus.groovy'; s(); def c = load 'scripts/groovy/registry-configure-nexus.groovy'; c(); e('nexus') }
                    if (params.PULP)  { def s = load 'scripts/groovy/install-pulp.groovy'; s(); e('pulp') }
                }
            }
        }

        stage('Git Servers') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'
                    if (params.GITEA)     { def s = load 'scripts/groovy/install-gitea.groovy'; s(); def c = load 'scripts/groovy/registry-configure-gitea.groovy'; c(); e('gitea') }
                    if (params.FORGEJO)   { def s = load 'scripts/groovy/install-forgejo.groovy'; s(); def c = load 'scripts/groovy/registry-configure-forgejo.groovy'; c(); e('forgejo') }
                    if (params.GOGS)      { def s = load 'scripts/groovy/install-gogs.groovy'; s(); e('gogs') }
                    if (params.GITBUCKET) { def s = load 'scripts/groovy/install-gitbucket.groovy'; s(); e('gitbucket') }
                    if (params.ONEDEV)    { def s = load 'scripts/groovy/install-onedev.groovy'; s(); e('onedev') }
                    if (params.GITLAB)    { def s = load 'scripts/groovy/install-gitlab.groovy'; s(); def c = load 'scripts/groovy/registry-configure-gitlab.groovy'; c(); e('gitlab') }
                }
            }
        }

        stage('Helm and Terraform Registries') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'
                    if (params.CHARTMUSEUM) { def s = load 'scripts/groovy/install-chartmuseum.groovy'; s(); def c = load 'scripts/groovy/registry-configure-chartmuseum.groovy'; c(); e('chartmuseum') }
                    if (params.TERRAREG)    { def s = load 'scripts/groovy/install-terrareg.groovy'; s(); e('terrareg') }
                }
            }
        }

        stage('Language Package Registries') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'
                    if (params.VERDACCIO)  { def s = load 'scripts/groovy/install-verdaccio.groovy'; s(); e('verdaccio') }
                    if (params.CNPMJS)     { def s = load 'scripts/groovy/install-cnpmjs.groovy'; s(); e('cnpmjs') }
                    if (params.PYPISERVER) { def s = load 'scripts/groovy/install-pypiserver.groovy'; s(); e('pypiserver') }
                    if (params.DEVPI)      { def s = load 'scripts/groovy/install-devpi.groovy'; s(); e('devpi') }
                    if (params.QUETZ)      { def s = load 'scripts/groovy/install-quetz.groovy'; s(); e('quetz') }
                    if (params.ATHENS)     { def s = load 'scripts/groovy/install-athens.groovy'; s(); e('athens') }
                    if (params.GOPROXY)    { def s = load 'scripts/groovy/install-goproxy.groovy'; s(); e('goproxy') }
                    if (params.REPOSILITE) { def s = load 'scripts/groovy/install-reposilite.groovy'; s(); e('reposilite') }
                    if (params.BAGET)      { def s = load 'scripts/groovy/install-baget.groovy'; s(); e('baget') }
                    if (params.KELLNR)     { def s = load 'scripts/groovy/install-kellnr.groovy'; s(); e('kellnr') }
                    if (params.ALEXANDRIE) { def s = load 'scripts/groovy/install-alexandrie.groovy'; s(); e('alexandrie') }
                    if (params.GEMINABOX)  { def s = load 'scripts/groovy/install-geminabox.groovy'; s(); e('geminabox') }
                    if (params.CONAN_SERVER){ def s = load 'scripts/groovy/install-conan-server.groovy'; s(); e('conan-server') }
                    if (params.APTLY)      { def s = load 'scripts/groovy/install-aptly.groovy'; s(); e('aptly') }
                }
            }
        }

        stage('Uninstall') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.APTLY)        { sh 'helm uninstall aptly -n aptly --ignore-not-found || true; kubectl delete pvc --all -n aptly --ignore-not-found || true; kubectl delete namespace aptly --ignore-not-found || true' }
                    if (params.CONAN_SERVER) { sh 'helm uninstall conan-server -n conan-server --ignore-not-found || true; kubectl delete namespace conan-server --ignore-not-found || true' }
                    if (params.GEMINABOX)    { sh 'helm uninstall geminabox -n geminabox --ignore-not-found || true; kubectl delete namespace geminabox --ignore-not-found || true' }
                    if (params.ALEXANDRIE)   { sh 'helm uninstall alexandrie -n alexandrie --ignore-not-found || true; kubectl delete pvc --all -n alexandrie --ignore-not-found || true; kubectl delete namespace alexandrie --ignore-not-found || true' }
                    if (params.KELLNR)       { sh 'helm uninstall kellnr -n kellnr --ignore-not-found || true; kubectl delete pvc --all -n kellnr --ignore-not-found || true; kubectl delete namespace kellnr --ignore-not-found || true' }
                    if (params.BAGET)        { sh 'helm uninstall baget -n baget --ignore-not-found || true; kubectl delete pvc --all -n baget --ignore-not-found || true; kubectl delete namespace baget --ignore-not-found || true' }
                    if (params.REPOSILITE)   { sh 'helm uninstall reposilite -n reposilite --ignore-not-found || true; kubectl delete pvc --all -n reposilite --ignore-not-found || true; kubectl delete namespace reposilite --ignore-not-found || true' }
                    if (params.GOPROXY)      { sh 'helm uninstall goproxy -n goproxy --ignore-not-found || true; kubectl delete namespace goproxy --ignore-not-found || true' }
                    if (params.ATHENS)       { sh 'helm uninstall athens -n athens --ignore-not-found || true; kubectl delete pvc --all -n athens --ignore-not-found || true; kubectl delete namespace athens --ignore-not-found || true' }
                    if (params.QUETZ)        { sh 'helm uninstall quetz -n quetz --ignore-not-found || true; kubectl delete namespace quetz --ignore-not-found || true' }
                    if (params.DEVPI)        { sh 'helm uninstall devpi -n devpi --ignore-not-found || true; kubectl delete pvc --all -n devpi --ignore-not-found || true; kubectl delete namespace devpi --ignore-not-found || true' }
                    if (params.PYPISERVER)   { sh 'helm uninstall pypiserver -n pypiserver --ignore-not-found || true; kubectl delete pvc --all -n pypiserver --ignore-not-found || true; kubectl delete namespace pypiserver --ignore-not-found || true' }
                    if (params.CNPMJS)       { sh 'helm uninstall cnpmjs -n cnpmjs --ignore-not-found || true; kubectl delete namespace cnpmjs --ignore-not-found || true' }
                    if (params.VERDACCIO)    { sh 'helm uninstall verdaccio -n verdaccio --ignore-not-found || true; kubectl delete pvc --all -n verdaccio --ignore-not-found || true; kubectl delete namespace verdaccio --ignore-not-found || true' }
                    if (params.TERRAREG)     { sh 'helm uninstall terrareg -n terrareg --ignore-not-found || true; kubectl delete namespace terrareg --ignore-not-found || true' }
                    if (params.CHARTMUSEUM)  { sh 'helm uninstall chartmuseum -n chartmuseum --ignore-not-found || true; kubectl delete namespace chartmuseum --ignore-not-found || true' }
                    if (params.GITLAB)       { sh 'helm uninstall gitlab -n gitlab --ignore-not-found || true; kubectl delete pvc --all -n gitlab --ignore-not-found || true; kubectl delete namespace gitlab --ignore-not-found || true' }
                    if (params.ONEDEV)       { sh 'helm uninstall onedev -n onedev --ignore-not-found || true; kubectl delete pvc --all -n onedev --ignore-not-found || true; kubectl delete namespace onedev --ignore-not-found || true' }
                    if (params.GITBUCKET)    { sh 'helm uninstall gitbucket -n gitbucket --ignore-not-found || true; kubectl delete pvc --all -n gitbucket --ignore-not-found || true; kubectl delete namespace gitbucket --ignore-not-found || true' }
                    if (params.GOGS)         { sh 'helm uninstall gogs -n gogs --ignore-not-found || true; kubectl delete pvc --all -n gogs --ignore-not-found || true; kubectl delete namespace gogs --ignore-not-found || true' }
                    if (params.FORGEJO)      { sh 'helm uninstall forgejo -n forgejo --ignore-not-found || true; kubectl delete pvc --all -n forgejo --ignore-not-found || true; kubectl delete namespace forgejo --ignore-not-found || true' }
                    if (params.GITEA)        { sh 'helm uninstall gitea -n gitea --ignore-not-found || true; kubectl delete pvc --all -n gitea --ignore-not-found || true; kubectl delete namespace gitea --ignore-not-found || true' }
                    if (params.PULP)         { sh 'helm uninstall pulp -n pulp --ignore-not-found || true; kubectl delete pvc --all -n pulp --ignore-not-found || true; kubectl delete namespace pulp --ignore-not-found || true' }
                    if (params.NEXUS)        { sh 'helm uninstall nexus -n nexus --ignore-not-found || true; kubectl delete pvc --all -n nexus --ignore-not-found || true; kubectl delete namespace nexus --ignore-not-found || true' }
                    if (params.DRAGONFLY)    { sh 'helm uninstall dragonfly -n dragonfly --ignore-not-found || true; kubectl delete pvc --all -n dragonfly --ignore-not-found || true; kubectl delete namespace dragonfly --ignore-not-found || true' }
                    if (params.KRAKEN)       { sh 'helm uninstall kraken -n kraken --ignore-not-found || true; kubectl delete namespace kraken --ignore-not-found || true' }
                    if (params.QUAY)         { sh 'helm uninstall quay -n quay --ignore-not-found || true; kubectl delete pvc --all -n quay --ignore-not-found || true; kubectl delete namespace quay --ignore-not-found || true' }
                    if (params.DISTRIBUTION) { sh 'helm uninstall distribution -n distribution --ignore-not-found || true; kubectl delete pvc --all -n distribution --ignore-not-found || true; kubectl delete namespace distribution --ignore-not-found || true' }
                    if (params.ZOT)          { sh 'helm uninstall zot -n zot --ignore-not-found || true; kubectl delete pvc --all -n zot --ignore-not-found || true; kubectl delete namespace zot --ignore-not-found || true' }
                    if (params.HARBOR)       { sh 'helm uninstall harbor -n harbor --ignore-not-found || true; kubectl delete pvc --all -n harbor --ignore-not-found || true; kubectl delete namespace harbor --ignore-not-found || true' }
                }
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
