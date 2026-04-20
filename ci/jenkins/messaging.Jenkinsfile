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
            description: 'INSTALL/UNINSTALL deploys messaging tools on Kubernetes. CONFIGURE applies post-install setup (topics, exchanges, streams, connectors).'
        )

        // ── Kafka Ecosystem ───────────────────────────────────────────────
        booleanParam(name: 'ZOOKEEPER',         defaultValue: false, description: 'Apache ZooKeeper — distributed coordination service for Kafka')
        booleanParam(name: 'KAFKA',             defaultValue: false, description: 'Apache Kafka — distributed event streaming platform (Confluent 7.7.1)')
        booleanParam(name: 'SCHEMA_REGISTRY',   defaultValue: false, description: 'Confluent Schema Registry — Avro/Protobuf/JSON schema management')
        booleanParam(name: 'KAFKA_CONNECT',     defaultValue: false, description: 'Kafka Connect — scalable data integration framework with connector ecosystem')
        booleanParam(name: 'KSQLDB',            defaultValue: false, description: 'ksqlDB — streaming SQL engine for real-time data processing on Kafka')
        booleanParam(name: 'STRIMZI',           defaultValue: false, description: 'Strimzi — Kubernetes operator for Apache Kafka cluster management')

        // ── Message Brokers ───────────────────────────────────────────────
        booleanParam(name: 'RABBITMQ',          defaultValue: false, description: 'RabbitMQ — reliable AMQP message broker with management UI (3.13)')
        booleanParam(name: 'NATS',              defaultValue: false, description: 'NATS JetStream — high performance cloud-native messaging (2.10)')
        booleanParam(name: 'PULSAR',            defaultValue: false, description: 'Apache Pulsar — cloud-native distributed messaging and streaming (3.3.0)')
        booleanParam(name: 'ACTIVEMQ_ARTEMIS',  defaultValue: false, description: 'ActiveMQ Artemis — next generation enterprise message broker')
        booleanParam(name: 'REDPANDA',          defaultValue: false, description: 'Redpanda — Kafka-compatible streaming platform without ZooKeeper')
        booleanParam(name: 'MEMPHIS',           defaultValue: false, description: 'Memphis — modern cloud-native message broker for developers')

        // ── Management UIs ────────────────────────────────────────────────
        booleanParam(name: 'KAFKA_UI',          defaultValue: false, description: 'Kafka UI — open source web UI for Kafka cluster management (Provectus)')
        booleanParam(name: 'AKHQ',              defaultValue: false, description: 'AKHQ — Kafka HQ management and monitoring dashboard')
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
                    if (!fileExists('infra.env')) {
                        error "infra.env not found — run the k8s-infra pipeline first to provision a cluster"
                    }
                    def kubeconfigContent = readFile('infra.env').trim()
                        .split('
').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!kubeconfigContent) {
                        error "KUBECONFIG_CONTENT missing from infra.env — run the k8s-infra pipeline first"
                    }
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kubeconfigContent
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        stage('Install Messaging Tools') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    // ZooKeeper must come before Kafka
                    if (params.ZOOKEEPER)       { def s = load 'scripts/groovy/messaging-install-zookeeper.groovy';       s() }
                    if (params.KAFKA)           { def s = load 'scripts/groovy/messaging-install-kafka.groovy';           s() }
                    if (params.SCHEMA_REGISTRY) { def s = load 'scripts/groovy/messaging-install-schema-registry.groovy'; s() }
                    if (params.KAFKA_CONNECT)   { def s = load 'scripts/groovy/messaging-install-kafka-connect.groovy';   s() }
                    if (params.KSQLDB)          { def s = load 'scripts/groovy/messaging-install-ksqldb.groovy';          s() }
                    if (params.STRIMZI)         { def s = load 'scripts/groovy/messaging-install-strimzi.groovy';         s() }
                    if (params.RABBITMQ)        { def s = load 'scripts/groovy/messaging-install-rabbitmq.groovy';        s() }
                    if (params.NATS)            { def s = load 'scripts/groovy/messaging-install-nats.groovy';            s() }
                    if (params.PULSAR)          { def s = load 'scripts/groovy/messaging-install-pulsar.groovy';          s() }
                    if (params.ACTIVEMQ_ARTEMIS){ def s = load 'scripts/groovy/messaging-install-activemq-artemis.groovy';s() }
                    if (params.REDPANDA)        { def s = load 'scripts/groovy/messaging-install-redpanda.groovy';        s() }
                    if (params.MEMPHIS)         { def s = load 'scripts/groovy/messaging-install-memphis.groovy';         s() }
                    if (params.KAFKA_UI)        { def s = load 'scripts/groovy/messaging-install-kafka-ui.groovy';        s() }
                    if (params.AKHQ)            { def s = load 'scripts/groovy/messaging-install-akhq.groovy';            s() }
                }
            }
        }

        stage('Uninstall Messaging Tools') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.AKHQ)            { sh 'helm uninstall akhq            -n akhq            --ignore-not-found || true' }
                    if (params.KAFKA_UI)        { sh 'helm uninstall kafka-ui        -n kafka-ui        --ignore-not-found || true' }
                    if (params.MEMPHIS)         { sh 'helm uninstall memphis         -n memphis         --ignore-not-found || true' }
                    if (params.REDPANDA)        { sh 'helm uninstall redpanda        -n redpanda        --ignore-not-found || true' }
                    if (params.ACTIVEMQ_ARTEMIS){ sh 'helm uninstall activemq-artemis -n activemq-artemis --ignore-not-found || true' }
                    if (params.PULSAR)          { sh 'helm uninstall pulsar          -n pulsar          --ignore-not-found || true' }
                    if (params.NATS)            { sh 'helm uninstall nats            -n nats            --ignore-not-found || true' }
                    if (params.RABBITMQ)        { sh 'helm uninstall rabbitmq        -n rabbitmq        --ignore-not-found || true' }
                    if (params.STRIMZI)         { sh 'helm uninstall strimzi         -n strimzi         --ignore-not-found || true' }
                    if (params.KSQLDB)          { sh 'helm uninstall ksqldb          -n ksqldb          --ignore-not-found || true' }
                    if (params.KAFKA_CONNECT)   { sh 'helm uninstall kafka-connect   -n kafka-connect   --ignore-not-found || true' }
                    if (params.SCHEMA_REGISTRY) { sh 'helm uninstall schema-registry -n schema-registry --ignore-not-found || true' }
                    if (params.KAFKA)           { sh 'helm uninstall kafka           -n kafka           --ignore-not-found || true' }
                    if (params.ZOOKEEPER)       { sh 'helm uninstall zookeeper       -n zookeeper       --ignore-not-found || true' }
                }
            }
        }

        stage('Configure Messaging Tools') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps {
                script {
                    if (params.KAFKA)           { def s = load 'scripts/groovy/messaging-configure-kafka.groovy';           s() }
                    if (params.SCHEMA_REGISTRY) { def s = load 'scripts/groovy/messaging-configure-schema-registry.groovy'; s() }
                    if (params.KAFKA_CONNECT)   { def s = load 'scripts/groovy/messaging-configure-kafka-connect.groovy';   s() }
                    if (params.RABBITMQ)        { def s = load 'scripts/groovy/messaging-configure-rabbitmq.groovy';        s() }
                    if (params.NATS)            { def s = load 'scripts/groovy/messaging-configure-nats.groovy';            s() }
                    if (params.REDPANDA)        { def s = load 'scripts/groovy/messaging-configure-redpanda.groovy';        s() }
                    if (params.MEMPHIS)         { def s = load 'scripts/groovy/messaging-configure-memphis.groovy';         s() }
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
