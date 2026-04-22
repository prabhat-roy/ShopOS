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
            choices: ['INSTALL', 'UNINSTALL'],
            description: 'INSTALL — deploy and configure all messaging tools. UNINSTALL — remove all.'
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
                        error "infra.env not found — run the k8s-infra pipeline first"
                    }
                    def content = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!content) error "KUBECONFIG_CONTENT missing from infra.env"
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: content
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        // ── INSTALL + CONFIGURE ───────────────────────────────────────────────

        stage('Zookeeper') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-zookeeper.groovy'; s()
                }
            }
        }

        stage('Kafka') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-kafka.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-kafka.groovy'; c()
                }
            }
        }

        stage('Schema Registry') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-schema-registry.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-schema-registry.groovy'; c()
                }
            }
        }

        stage('Kafka Connect') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-kafka-connect.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-kafka-connect.groovy'; c()
                }
            }
        }

        stage('ksqlDB') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-ksqldb.groovy'; s()
                }
            }
        }

        stage('Strimzi') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-strimzi.groovy'; s()
                }
            }
        }

        stage('RabbitMQ') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-rabbitmq.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-rabbitmq.groovy'; c()
                }
            }
        }

        stage('NATS') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-nats.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-nats.groovy'; c()
                }
            }
        }

        stage('Pulsar') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-pulsar.groovy'; s()
                }
            }
        }

        stage('ActiveMQ Artemis') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-activemq-artemis.groovy'; s()
                }
            }
        }

        stage('Redpanda') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-redpanda.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-redpanda.groovy'; c()
                }
            }
        }

        stage('Memphis') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-memphis.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-memphis.groovy'; c()
                }
            }
        }

        stage('Kafka UI') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-kafka-ui.groovy'; s()
                }
            }
        }

        stage('AKHQ') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-akhq.groovy'; s()
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

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
        success { echo "${params.ACTION} completed successfully." }
        failure { echo "${params.ACTION} failed — check stage logs above." }
        cleanup { sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true" }
    }
}
