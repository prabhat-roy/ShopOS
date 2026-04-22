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
            choices: ['INSTALL', 'UNINSTALL'],
            description: 'INSTALL — deploy selected streaming tools. UNINSTALL — remove selected.'
        )
        booleanParam(name: 'FLINK',             defaultValue: true, description: 'Apache Flink — real-time stream processing operator + jobs')
        booleanParam(name: 'DEBEZIUM_POSTGRES', defaultValue: true, description: 'Debezium Postgres CDC — change data capture from PostgreSQL')
        booleanParam(name: 'DEBEZIUM_MONGODB',  defaultValue: true, description: 'Debezium MongoDB CDC — change data capture from MongoDB')
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
                    if (!fileExists('infra.env')) error "infra.env not found — run k8s-infra pipeline first"
                    def content = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!content) error "KUBECONFIG_CONTENT missing from infra.env"
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: content
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        // ── INSTALL ──────────────────────────────────────────────────────────

        stage('Flink Operator') {
            when { expression { params.ACTION == 'INSTALL' && params.FLINK } }
            steps {
                sh """
                    kubectl create namespace flink-system --dry-run=client -o yaml | kubectl apply -f -
                    helm repo add flink-operator https://downloads.apache.org/flink/flink-kubernetes-operator-1.9.0/ || true
                    helm repo update
                    helm upgrade --install flink-operator flink-operator/flink-kubernetes-operator \
                        --namespace flink-system \
                        --set webhook.create=false \
                        --wait --timeout=5m
                    echo "Flink Operator installed"
                """
            }
        }

        stage('Flink Jobs') {
            when { expression { params.ACTION == 'INSTALL' && params.FLINK } }
            steps {
                sh """
                    kubectl create namespace streaming --dry-run=client -o yaml | kubectl apply -f -
                    kubectl apply -f streaming/flink/order-analytics-job.yaml -n streaming
                    kubectl apply -f streaming/flink/fraud-detection-job.yaml -n streaming
                    kubectl rollout status deployment/flink-jobmanager -n streaming --timeout=120s || true
                    echo "Flink jobs submitted"
                """
            }
        }

        stage('Debezium — Postgres CDC') {
            when { expression { params.ACTION == 'INSTALL' && params.DEBEZIUM_POSTGRES } }
            steps {
                sh """
                    # Wait for Kafka Connect to be available
                    KAFKA_CONNECT_URL=\$(kubectl get svc kafka-connect -n shopos-infra -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo 'kafka-connect.shopos-infra.svc.cluster.local')
                    echo "Registering Postgres CDC connector..."
                    kubectl run debezium-reg-postgres --image=curlimages/curl:8.10.1 --restart=Never \
                        --env="KAFKA_CONNECT_URL=\${KAFKA_CONNECT_URL}" \
                        --command -- /bin/sh -c "
                            sleep 10
                            curl -sf -X POST http://\${KAFKA_CONNECT_URL}:8083/connectors \
                              -H 'Content-Type: application/json' \
                              -d @/dev/stdin < /streaming/debezium/postgres-orders-connector.json || echo 'Connector may already exist'
                        " 2>/dev/null || true
                    echo "Postgres CDC connector registered"
                """
            }
        }

        stage('Debezium — MongoDB CDC') {
            when { expression { params.ACTION == 'INSTALL' && params.DEBEZIUM_MONGODB } }
            steps {
                sh """
                    KAFKA_CONNECT_URL=\$(kubectl get svc kafka-connect -n shopos-infra -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo 'kafka-connect.shopos-infra.svc.cluster.local')
                    echo "Registering MongoDB CDC connector..."
                    kubectl run debezium-reg-mongo --image=curlimages/curl:8.10.1 --restart=Never \
                        --env="KAFKA_CONNECT_URL=\${KAFKA_CONNECT_URL}" \
                        --command -- /bin/sh -c "
                            sleep 10
                            curl -sf -X POST http://\${KAFKA_CONNECT_URL}:8083/connectors \
                              -H 'Content-Type: application/json' \
                              -d @/dev/stdin < /streaming/debezium/mongodb-catalog-connector.json || echo 'Connector may already exist'
                        " 2>/dev/null || true
                    echo "MongoDB CDC connector registered"
                """
            }
        }

        stage('Verify Streaming') {
            when { expression { params.ACTION == 'INSTALL' && params.DEBEZIUM_MONGODB } }
            steps {
                sh """
                    echo "=== Flink deployments ==="
                    kubectl get flinkdeployment -n streaming 2>/dev/null || kubectl get pods -n streaming
                    echo "=== Kafka Connect connectors ==="
                    KAFKA_CONNECT_URL=\$(kubectl get svc kafka-connect -n shopos-infra -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo 'kafka-connect.shopos-infra.svc.cluster.local')
                    curl -sf http://\${KAFKA_CONNECT_URL}:8083/connectors 2>/dev/null || echo "Connect not yet available — connectors queued"
                """
            }
        }

        // ── UNINSTALL ────────────────────────────────────────────────────────

        stage('Uninstall Flink Jobs') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FLINK } }
            steps {
                sh """
                    kubectl delete -f streaming/flink/order-analytics-job.yaml -n streaming --ignore-not-found
                    kubectl delete -f streaming/flink/fraud-detection-job.yaml -n streaming --ignore-not-found
                    echo "Flink jobs removed"
                """
            }
        }

        stage('Uninstall Debezium Connectors') {
            when { expression { params.ACTION == 'UNINSTALL' && params.DEBEZIUM_POSTGRES } }
            steps {
                sh """
                    KAFKA_CONNECT_URL=\$(kubectl get svc kafka-connect -n shopos-infra -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo 'kafka-connect.shopos-infra.svc.cluster.local')
                    curl -sf -X DELETE http://\${KAFKA_CONNECT_URL}:8083/connectors/postgres-orders-connector || true
                    curl -sf -X DELETE http://\${KAFKA_CONNECT_URL}:8083/connectors/mongodb-catalog-connector || true
                    echo "Debezium connectors removed"
                """
            }
        }

        stage('Uninstall Flink Operator') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FLINK } }
            steps {
                sh "helm uninstall flink-operator -n flink-system --ignore-not-found || true"
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
        success { echo "${params.ACTION} of streaming pipeline completed successfully." }
        failure { echo "${params.ACTION} failed — check stage logs above." }
    }
}
