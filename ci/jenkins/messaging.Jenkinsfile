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
            description: 'INSTALL — deploy, configure and apply K8s enhancements for all messaging tools. UNINSTALL — remove all.'
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

        // ── INSTALL + CONFIGURE + K8s ENHANCEMENTS ───────────────────────────

        stage('Zookeeper') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-zookeeper.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zookeeper')
                }
            }
        }

        stage('Kafka') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-kafka.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-kafka.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kafka')
                }
            }
        }

        stage('Schema Registry') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-schema-registry.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-schema-registry.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('schema-registry')
                }
            }
        }

        stage('Kafka Connect') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-kafka-connect.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-kafka-connect.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kafka-connect')
                }
            }
        }

        stage('ksqlDB') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-ksqldb.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('ksqldb')
                }
            }
        }

        stage('Strimzi') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-strimzi.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('strimzi')
                }
            }
        }

        stage('RabbitMQ') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-rabbitmq.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-rabbitmq.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('rabbitmq')
                }
            }
        }

        stage('NATS') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-nats.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-nats.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('nats')
                }
            }
        }

        stage('Pulsar') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-pulsar.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pulsar')
                }
            }
        }

        stage('ActiveMQ Artemis') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-activemq-artemis.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('activemq-artemis')
                }
            }
        }

        stage('Redpanda') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-redpanda.groovy'; s()
                    def c = load 'scripts/groovy/messaging-configure-redpanda.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('redpanda')
                }
            }
        }

        stage('Memphis') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    catchError(buildResult: 'SUCCESS', stageResult: 'UNSTABLE') {
                        def s = load 'scripts/groovy/messaging-install-memphis.groovy'; s()
                        def c = load 'scripts/groovy/messaging-configure-memphis.groovy'; c()
                        def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('memphis')
                    }
                }
            }
        }

        stage('Kafka UI') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-kafka-ui.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kafka-ui')
                }
            }
        }

        stage('AKHQ') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/messaging-install-akhq.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('akhq')
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall AKHQ') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall akhq -n akhq --ignore-not-found || true
                    kubectl delete namespace akhq --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kafka UI') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall kafka-ui -n kafka-ui --ignore-not-found || true
                    kubectl delete namespace kafka-ui --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Memphis') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall memphis -n memphis --ignore-not-found || true
                    kubectl delete pvc --all -n memphis --ignore-not-found || true
                    kubectl delete namespace memphis --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Redpanda') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall redpanda -n redpanda --ignore-not-found || true
                    kubectl delete pvc --all -n redpanda --ignore-not-found || true
                    kubectl delete namespace redpanda --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall ActiveMQ Artemis') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall activemq-artemis -n activemq-artemis --ignore-not-found || true
                    kubectl delete pvc --all -n activemq-artemis --ignore-not-found || true
                    kubectl delete namespace activemq-artemis --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pulsar') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall pulsar -n pulsar --ignore-not-found || true
                    kubectl delete pvc --all -n pulsar --ignore-not-found || true
                    kubectl delete namespace pulsar --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall NATS') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall nats -n nats --ignore-not-found || true
                    kubectl delete pvc --all -n nats --ignore-not-found || true
                    kubectl delete namespace nats --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall RabbitMQ') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall rabbitmq -n rabbitmq --ignore-not-found || true
                    kubectl delete pvc --all -n rabbitmq --ignore-not-found || true
                    kubectl delete namespace rabbitmq --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Strimzi') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall strimzi -n strimzi --ignore-not-found || true
                    kubectl delete namespace strimzi --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall ksqlDB') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall ksqldb -n ksqldb --ignore-not-found || true
                    kubectl delete namespace ksqldb --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kafka Connect') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall kafka-connect -n kafka-connect --ignore-not-found || true
                    kubectl delete namespace kafka-connect --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Schema Registry') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall schema-registry -n schema-registry --ignore-not-found || true
                    kubectl delete namespace schema-registry --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kafka') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall kafka -n kafka --ignore-not-found || true
                    kubectl delete pvc --all -n kafka --ignore-not-found || true
                    kubectl delete namespace kafka --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Zookeeper') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    helm uninstall zookeeper -n zookeeper --ignore-not-found || true
                    kubectl delete pvc --all -n zookeeper --ignore-not-found || true
                    kubectl delete namespace zookeeper --ignore-not-found || true
                '''
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
