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
    }

    stages {

        stage('Git Fetch') {
            steps {
                checkout scm
                sh 'test -f /var/lib/jenkins/infra.env && cp /var/lib/jenkins/infra.env . || true'
            }
        }

        // ── NETWORKING ──────────────────────────────────────────────────────
        // Step 1 of 3: Install CNI first — nothing works without a network plugin.
        // Step 2 of 3: Install Ingress + Service Mesh (after cert-manager from Security Step 1).
        // Step 3 of 3: Configure mTLS, namespace injection, IngressClass, ExternalDNS.

        stage('Networking — Install CNI') {
            when { expression { !params.SKIP_NETWORKING } }
            steps {
                echo '=== Networking Step 1/3: CNI ==='
                build job: 'networking',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',       value: 'INSTALL'),
                        booleanParam(name: 'CILIUM',  value: true),
                    ]
            }
        }

        stage('Security — Install TLS Foundation') {
            when { expression { !params.SKIP_SECURITY } }
            steps {
                echo '=== Security Step 1/4: cert-manager (required by Vault, Istio, Keycloak) ==='
                build job: 'security',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',            value: 'INSTALL'),
                        booleanParam(name: 'CERT_MANAGER', value: true),
                    ]
            }
        }

        stage('Networking — Install Ingress + Service Mesh') {
            when { expression { !params.SKIP_NETWORKING } }
            steps {
                echo '=== Networking Step 2/3: Traefik + Istio + ExternalDNS ==='
                build job: 'networking',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',          value: 'INSTALL'),
                        booleanParam(name: 'TRAEFIK',    value: true),
                        booleanParam(name: 'ISTIO',      value: true),
                        booleanParam(name: 'EXTERNAL_DNS', value: true),
                    ]
            }
        }

        stage('Security — Install Secrets + Identity + Policy') {
            when { expression { !params.SKIP_SECURITY } }
            steps {
                echo '=== Security Step 2/4: Vault + Keycloak + ESO + Kyverno + OPA ==='
                build job: 'security',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',                   value: 'INSTALL'),
                        booleanParam(name: 'VAULT',               value: true),
                        booleanParam(name: 'KEYCLOAK',            value: true),
                        booleanParam(name: 'EXTERNAL_SECRETS',    value: true),
                        booleanParam(name: 'KYVERNO',             value: true),
                        booleanParam(name: 'OPA',                 value: true),
                        booleanParam(name: 'SPIRE',               value: true),
                    ]
            }
        }

        stage('Security — Configure Vault + Keycloak + Policies') {
            when { expression { !params.SKIP_SECURITY } }
            steps {
                echo '=== Security Step 3/4: Configure Vault PKI, Keycloak realm, OPA policies, Kyverno ClusterPolicies ==='
                build job: 'security',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',                value: 'CONFIGURE'),
                        booleanParam(name: 'VAULT',            value: true),
                        booleanParam(name: 'KEYCLOAK',         value: true),
                        booleanParam(name: 'EXTERNAL_SECRETS', value: true),
                        booleanParam(name: 'OPA',              value: true),
                        booleanParam(name: 'KYVERNO',          value: true),
                    ]
            }
        }

        stage('Networking — Configure mTLS + Ingress + DNS') {
            when { expression { !params.SKIP_NETWORKING } }
            steps {
                echo '=== Networking Step 3/3: Istio namespace injection, Traefik IngressClass, ExternalDNS provider ==='
                build job: 'networking',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',            value: 'CONFIGURE'),
                        booleanParam(name: 'ISTIO',        value: true),
                        booleanParam(name: 'TRAEFIK',      value: true),
                        booleanParam(name: 'EXTERNAL_DNS', value: true),
                    ]
            }
        }

        stage('Security — Install Runtime Threat Detection') {
            when { expression { !params.SKIP_SECURITY } }
            steps {
                echo '=== Security Step 4/4: Falco + Tetragon (monitoring running pods — must be last) ==='
                build job: 'security',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',           value: 'INSTALL'),
                        booleanParam(name: 'FALCO',       value: true),
                        booleanParam(name: 'TETRAGON',    value: true),
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
                        string(name: 'ACTION',                   value: 'INSTALL'),
                        booleanParam(name: 'PROMETHEUS',          value: true),
                        booleanParam(name: 'GRAFANA',             value: true),
                        booleanParam(name: 'ALERTMANAGER',        value: true),
                        booleanParam(name: 'LOKI',                value: true),
                        booleanParam(name: 'JAEGER',              value: true),
                        booleanParam(name: 'OTEL_COLLECTOR',      value: true),
                        booleanParam(name: 'FLUENT_BIT',          value: true),
                        booleanParam(name: 'KUBE_STATE_METRICS',  value: true),
                        booleanParam(name: 'NODE_EXPORTER',       value: true),
                    ]
            }
        }

        stage('Observability — Configure') {
            when { expression { !params.SKIP_OBSERVABILITY } }
            steps {
                echo '=== Observability Configure: datasources, alert rules, OTel scrape configs ==='
                build job: 'observability',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',               value: 'CONFIGURE'),
                        booleanParam(name: 'PROMETHEUS',      value: true),
                        booleanParam(name: 'GRAFANA',         value: true),
                        booleanParam(name: 'ALERTMANAGER',    value: true),
                        booleanParam(name: 'LOKI',            value: true),
                        booleanParam(name: 'JAEGER',          value: true),
                        booleanParam(name: 'OTEL_COLLECTOR',  value: true),
                        booleanParam(name: 'FLUENT_BIT',      value: true),
                    ]
            }
        }

        // ── MESSAGING ───────────────────────────────────────────────────────
        // Services cannot start without Kafka + Schema Registry.
        // This is the final bootstrap step before any service is deployed.

        stage('Messaging — Install') {
            when { expression { !params.SKIP_MESSAGING } }
            steps {
                echo '=== Messaging: Kafka + ZooKeeper + Schema Registry + RabbitMQ + NATS ==='
                build job: 'messaging',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',                   value: 'INSTALL'),
                        booleanParam(name: 'KAFKA',               value: true),
                        booleanParam(name: 'ZOOKEEPER',           value: true),
                        booleanParam(name: 'SCHEMA_REGISTRY',     value: true),
                        booleanParam(name: 'RABBITMQ',            value: true),
                        booleanParam(name: 'NATS',                value: true),
                        booleanParam(name: 'AKHQ',                value: true),
                    ]
            }
        }

        stage('Messaging — Configure') {
            when { expression { !params.SKIP_MESSAGING } }
            steps {
                echo '=== Messaging Configure: Kafka topics, Avro schemas, RabbitMQ vhosts ==='
                build job: 'messaging',
                    wait: true,
                    parameters: [
                        string(name: 'ACTION',               value: 'CONFIGURE'),
                        booleanParam(name: 'KAFKA',           value: true),
                        booleanParam(name: 'SCHEMA_REGISTRY', value: true),
                        booleanParam(name: 'RABBITMQ',        value: true),
                        booleanParam(name: 'NATS',            value: true),
                    ]
            }
        }

        stage('Bootstrap Complete') {
            steps {
                echo """
==========================================================
  Cluster bootstrap complete for environment: ${params.ENVIRONMENT}

  Infrastructure ready:
    CNI        : Cilium
    Ingress    : Traefik
    Mesh       : Istio (mTLS enabled on all domain namespaces)
    DNS        : ExternalDNS
    TLS        : cert-manager
    Secrets    : Vault + External Secrets Operator
    IAM        : Keycloak
    Policy     : Kyverno + OPA
    Runtime    : Falco + Tetragon
    Metrics    : Prometheus + Grafana + Alertmanager
    Logs       : Loki + Fluent Bit
    Tracing    : Jaeger + OTel Collector
    Messaging  : Kafka + Schema Registry + RabbitMQ + NATS

  Next step: trigger deploy.Jenkinsfile to deploy services.
==========================================================
                """
            }
        }
    }

    post {
        failure {
            echo 'Bootstrap failed — check the failed stage above. Fix the issue and re-run with SKIP_* flags to resume from where it stopped.'
        }
        cleanup {
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
    }
}
