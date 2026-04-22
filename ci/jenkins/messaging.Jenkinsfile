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
            choices: ['INSTALL', 'UNINSTALL', 'CONFIGURE'],
            description: 'INSTALL — deploy all messaging tools on Kubernetes. UNINSTALL — remove all. CONFIGURE — post-install setup (topics, exchanges, streams, connectors).'
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

        // ── INSTALL ──────────────────────────────────────────────────────────

        stage('Install Zookeeper') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-zookeeper.groovy'; s() } }
        }

        stage('Install Kafka') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-kafka.groovy'; s() } }
        }

        stage('Install Schema Registry') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-schema-registry.groovy'; s() } }
        }

        stage('Install Kafka Connect') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-kafka-connect.groovy'; s() } }
        }

        stage('Install ksqlDB') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-ksqldb.groovy'; s() } }
        }

        stage('Install Strimzi') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-strimzi.groovy'; s() } }
        }

        stage('Install RabbitMQ') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-rabbitmq.groovy'; s() } }
        }

        stage('Install NATS') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-nats.groovy'; s() } }
        }

        stage('Install Pulsar') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-pulsar.groovy'; s() } }
        }

        stage('Install ActiveMQ Artemis') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-activemq-artemis.groovy'; s() } }
        }

        stage('Install Redpanda') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-redpanda.groovy'; s() } }
        }

        stage('Install Memphis') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-memphis.groovy'; s() } }
        }

        stage('Install Kafka UI') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-kafka-ui.groovy'; s() } }
        }

        stage('Install AKHQ') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps { script { def s = load 'scripts/groovy/messaging-install-akhq.groovy'; s() } }
        }

        // ── CONFIGURE ────────────────────────────────────────────────────────

        stage('Configure Kafka') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-kafka.groovy'; s() } }
        }

        stage('Configure Schema Registry') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-schema-registry.groovy'; s() } }
        }

        stage('Configure Kafka Connect') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-kafka-connect.groovy'; s() } }
        }

        stage('Configure RabbitMQ') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-rabbitmq.groovy'; s() } }
        }

        stage('Configure NATS') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-nats.groovy'; s() } }
        }

        stage('Configure Redpanda') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-redpanda.groovy'; s() } }
        }

        stage('Configure Memphis') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps { script { def s = load 'scripts/groovy/messaging-configure-memphis.groovy'; s() } }
        }

        // ── UNINSTALL (reverse install order) ────────────────────────────────

        stage('Uninstall AKHQ') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall akhq -n akhq --ignore-not-found || true' }
        }

        stage('Uninstall Kafka UI') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kafka-ui -n kafka-ui --ignore-not-found || true' }
        }

        stage('Uninstall Memphis') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall memphis -n memphis --ignore-not-found || true' }
        }

        stage('Uninstall Redpanda') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall redpanda -n redpanda --ignore-not-found || true' }
        }

        stage('Uninstall ActiveMQ Artemis') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall activemq-artemis -n activemq-artemis --ignore-not-found || true' }
        }

        stage('Uninstall Pulsar') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall pulsar -n pulsar --ignore-not-found || true' }
        }

        stage('Uninstall NATS') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall nats -n nats --ignore-not-found || true' }
        }

        stage('Uninstall RabbitMQ') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall rabbitmq -n rabbitmq --ignore-not-found || true' }
        }

        stage('Uninstall Strimzi') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall strimzi -n strimzi --ignore-not-found || true' }
        }

        stage('Uninstall ksqlDB') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall ksqldb -n ksqldb --ignore-not-found || true' }
        }

        stage('Uninstall Kafka Connect') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kafka-connect -n kafka-connect --ignore-not-found || true' }
        }

        stage('Uninstall Schema Registry') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall schema-registry -n schema-registry --ignore-not-found || true' }
        }

        stage('Uninstall Kafka') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kafka -n kafka --ignore-not-found || true' }
        }

        stage('Uninstall Zookeeper') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall zookeeper -n zookeeper --ignore-not-found || true' }
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
        }
    }
}
