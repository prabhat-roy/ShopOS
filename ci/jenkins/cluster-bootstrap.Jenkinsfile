pipeline {
    agent any

    options {
        timestamps()
        ansiColor('xterm')
        buildDiscarder(logRotator(numToKeepStr: '5'))
        timeout(time: 180, unit: 'MINUTES')
    }

    parameters {
        choice(
            name: 'ENVIRONMENT',
            choices: ['dev', 'staging', 'prod'],
            description: 'Target cluster environment'
        )
        booleanParam(
            name: 'SKIP_NETWORKING',
            defaultValue: false,
            description: 'Skip networking bootstrap (CNI + Ingress + Mesh) — use if already installed'
        )
        booleanParam(
            name: 'SKIP_SECURITY',
            defaultValue: false,
            description: 'Skip security bootstrap (Vault + Keycloak + Kyverno + Falco) — use if already installed'
        )
        booleanParam(
            name: 'SKIP_OBSERVABILITY',
            defaultValue: false,
            description: 'Skip observability bootstrap — use if already installed'
        )
        booleanParam(
            name: 'SKIP_MESSAGING',
            defaultValue: false,
            description: 'Skip messaging bootstrap (Kafka + RabbitMQ + NATS) — use if already installed'
        )
        booleanParam(
            name: 'SKIP_GITOPS',
            defaultValue: false,
            description: 'Skip GitOps bootstrap (ArgoCD + Flux + Argo Workflows) — use if already installed'
        )
        booleanParam(
            name: 'INSTALL_CROSSPLANE',
            defaultValue: false,
            description: 'Install Crossplane for Kubernetes-native IaC resource management'
        )
        booleanParam(
            name: 'SKIP_DATABASES',
            defaultValue: false,
            description: 'Skip specialist databases (ClickHouse, Weaviate, Neo4j, ScyllaDB, Temporal)'
        )
        booleanParam(
            name: 'SKIP_STREAMING',
            defaultValue: false,
            description: 'Skip streaming bootstrap (Flink operator + Debezium CDC connectors)'
        )
        booleanParam(
            name: 'SKIP_REGISTRY',
            defaultValue: false,
            description: 'Skip registry bootstrap (Harbor + Nexus + Gitea + ChartMuseum)'
        )
    }

    stages {

        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        // ── NETWORKING ──────────────────────────────────────────────────────
        // Installs all CNI, ingress, service mesh, and DNS tools in order.

        stage('Networking — Install') {
            when { expression { !params.SKIP_NETWORKING } }
            steps {
                echo '=== Networking: CNI + Ingress + Service Mesh + DNS ==='
                build job: 'networking',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── SECURITY ────────────────────────────────────────────────────────
        // Installs cert-manager, Vault, Keycloak, OPA, Kyverno, Falco, etc.

        stage('Security — Install') {
            when { expression { !params.SKIP_SECURITY } }
            steps {
                echo '=== Security: cert-manager + Vault + Keycloak + ESO + Kyverno + OPA + Falco ==='
                build job: 'security',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── OBSERVABILITY ───────────────────────────────────────────────────
        // Deploy before services so metrics/traces are captured from day 1.

        stage('Observability — Install') {
            when { expression { !params.SKIP_OBSERVABILITY } }
            steps {
                echo '=== Observability: Prometheus + Grafana + Loki + Jaeger + OTel + Alertmanager ==='
                build job: 'observability',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── MESSAGING ───────────────────────────────────────────────────────
        // Services cannot start without Kafka + Schema Registry.

        stage('Messaging — Install') {
            when { expression { !params.SKIP_MESSAGING } }
            steps {
                echo '=== Messaging: Kafka + ZooKeeper + Schema Registry + RabbitMQ + NATS ==='
                build job: 'messaging',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── GITOPS ──────────────────────────────────────────────────────────
        // Install after infrastructure is ready so ArgoCD can manage services.

        stage('GitOps — Install') {
            when { expression { !params.SKIP_GITOPS } }
            steps {
                echo '=== GitOps: ArgoCD + Flux + Argo Workflows + Argo Events + Sealed Secrets ==='
                build job: 'gitops',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── REGISTRY ────────────────────────────────────────────────────────
        // Harbor + Nexus + Gitea for artifact storage before builds run.

        stage('Registry — Install') {
            when { expression { !params.SKIP_REGISTRY } }
            steps {
                echo '=== Registry: Harbor + Nexus + Gitea + ChartMuseum + Zot ==='
                build job: 'registry',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── DATABASES ───────────────────────────────────────────────────────
        // ClickHouse, Weaviate, Neo4j, ScyllaDB, Temporal.

        stage('Databases — Install') {
            when { expression { !params.SKIP_DATABASES } }
            steps {
                echo '=== Databases: ClickHouse + Weaviate + Neo4j + ScyllaDB + Temporal ==='
                build job: 'databases',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        // ── STREAMING ───────────────────────────────────────────────────────
        // Flink operator + Debezium CDC connectors (after Kafka is running).

        stage('Streaming — Install') {
            when { expression { !params.SKIP_STREAMING } }
            steps {
                echo '=== Streaming: Flink Operator + Debezium Postgres/MongoDB CDC ==='
                build job: 'streaming',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION', value: 'INSTALL')
                    ]
            }
        }

        stage('Crossplane — Install') {
            when { expression { params.INSTALL_CROSSPLANE } }
            steps {
                script {
                    def s = load 'scripts/groovy/install-crossplane.groovy'
                    s()
                }
            }
        }

        stage('Bootstrap Complete') {
            steps {
                echo """
==========================================================
  Cluster bootstrap complete for environment: ${params.ENVIRONMENT}

  Infrastructure ready:
    Networking   : Cilium + Traefik + Istio + ExternalDNS
    Security     : cert-manager + Vault + Keycloak + Kyverno + OPA + Falco
    Observability: Prometheus + Grafana + Loki + Jaeger + OTel Collector
    Messaging    : Kafka + Schema Registry + RabbitMQ + NATS
    GitOps       : ArgoCD + Flux + Argo Workflows + Argo Events
    Registry     : Harbor + Nexus + Gitea + ChartMuseum
    Databases    : ClickHouse + Weaviate + Neo4j + ScyllaDB + Temporal
    Streaming    : Flink + Debezium (Postgres + MongoDB CDC)

  Next step: trigger deploy.Jenkinsfile to deploy all 224 services.
==========================================================
                """
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
        }
        failure {
            echo 'Bootstrap failed — check the failed stage above. Fix the issue and re-run with SKIP_* flags to resume from where it stopped.'
        }
        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
    }
}
