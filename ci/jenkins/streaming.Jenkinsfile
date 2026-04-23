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
        booleanParam(name: 'FLINK',             defaultValue: false, description: 'Apache Flink — real-time stream processing operator + jobs')
        booleanParam(name: 'DEBEZIUM_POSTGRES', defaultValue: false, description: 'Debezium Postgres CDC — change data capture from PostgreSQL')
        booleanParam(name: 'DEBEZIUM_MONGODB',  defaultValue: false, description: 'Debezium MongoDB CDC — change data capture from MongoDB')
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
                    kubectl create namespace analytics-ai --dry-run=client -o yaml | kubectl apply -f -
                    helm upgrade --install flink-operator streaming/flink/charts/flink-kubernetes-operator \
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
                    # Create flink ServiceAccount in analytics-ai namespace
                    kubectl create serviceaccount flink -n analytics-ai --dry-run=client -o yaml | kubectl apply -f -

                    # Apply FlinkDeployment CRDs — namespace comes from the YAML (analytics-ai)
                    kubectl apply -f streaming/flink/order-analytics-job.yaml
                    kubectl apply -f streaming/flink/fraud-detection-job.yaml

                    # Wait for jobmanager deployments created by the Flink operator
                    kubectl rollout status deployment/order-analytics-job -n analytics-ai --timeout=120s || true
                    kubectl rollout status deployment/fraud-detection-job  -n analytics-ai --timeout=120s || true
                    echo "Flink jobs submitted"
                """
            }
        }

        stage('Debezium — Postgres CDC') {
            when { expression { params.ACTION == 'INSTALL' && params.DEBEZIUM_POSTGRES } }
            steps {
                sh """
                    echo "Registering Postgres CDC connector via port-forward..."
                    kubectl port-forward svc/kafka-connect 18083:8083 -n shopos-infra &
                    PF_PID=\$!
                    sleep 8
                    curl -sf -X POST http://localhost:18083/connectors \
                        -H 'Content-Type: application/json' \
                        -d @streaming/debezium/postgres-orders-connector.json \
                        || echo 'Connector may already exist'
                    kill \$PF_PID 2>/dev/null || true
                    echo "Postgres CDC connector registered"
                """
            }
        }

        stage('Debezium — MongoDB CDC') {
            when { expression { params.ACTION == 'INSTALL' && params.DEBEZIUM_MONGODB } }
            steps {
                sh """
                    echo "Registering MongoDB CDC connector via port-forward..."
                    kubectl port-forward svc/kafka-connect 18083:8083 -n shopos-infra &
                    PF_PID=\$!
                    sleep 8
                    curl -sf -X POST http://localhost:18083/connectors \
                        -H 'Content-Type: application/json' \
                        -d @streaming/debezium/mongodb-catalog-connector.json \
                        || echo 'Connector may already exist'
                    kill \$PF_PID 2>/dev/null || true
                    echo "MongoDB CDC connector registered"
                """
            }
        }

        stage('Verify Streaming') {
            when { expression { params.ACTION == 'INSTALL' && (params.FLINK || params.DEBEZIUM_POSTGRES || params.DEBEZIUM_MONGODB) } }
            steps {
                sh """
                    if ${params.FLINK}; then
                        echo "=== Flink deployments (analytics-ai) ==="
                        kubectl get flinkdeployment -n analytics-ai 2>/dev/null || kubectl get pods -n analytics-ai
                    fi

                    if ${params.DEBEZIUM_POSTGRES} || ${params.DEBEZIUM_MONGODB}; then
                        echo "=== Kafka Connect connectors ==="
                        kubectl port-forward svc/kafka-connect 18083:8083 -n shopos-infra &
                        PF_PID=\$!
                        sleep 8
                        curl -sf http://localhost:18083/connectors 2>/dev/null \
                            || echo "Connect not yet available — connectors queued"
                        kill \$PF_PID 2>/dev/null || true
                    fi
                """
            }
        }

        // ── UNINSTALL ────────────────────────────────────────────────────────

        stage('Uninstall Flink Jobs') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FLINK } }
            steps {
                sh """
                    # Namespace comes from the YAML (analytics-ai) — no -n flag override
                    kubectl delete -f streaming/flink/order-analytics-job.yaml --ignore-not-found
                    kubectl delete -f streaming/flink/fraud-detection-job.yaml --ignore-not-found
                    echo "Flink jobs removed"
                """
            }
        }

        stage('Uninstall Debezium Connectors') {
            when { expression { params.ACTION == 'UNINSTALL' && (params.DEBEZIUM_POSTGRES || params.DEBEZIUM_MONGODB) } }
            steps {
                sh """
                    kubectl port-forward svc/kafka-connect 18083:8083 -n shopos-infra &
                    PF_PID=\$!
                    sleep 8
                    if ${params.DEBEZIUM_POSTGRES}; then
                        curl -sf -X DELETE http://localhost:18083/connectors/postgres-orders-connector || true
                        echo "Postgres CDC connector removed"
                    fi
                    if ${params.DEBEZIUM_MONGODB}; then
                        curl -sf -X DELETE http://localhost:18083/connectors/mongodb-catalog-connector || true
                        echo "MongoDB CDC connector removed"
                    fi
                    kill \$PF_PID 2>/dev/null || true
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
