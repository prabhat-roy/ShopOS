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
            description: 'INSTALL — deploy all specialist databases. UNINSTALL — remove all.'
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
                    if (!fileExists('infra.env')) error "infra.env not found — run k8s-infra pipeline first"
                    def content = readFile('infra.env').trim()
                        .split('\n').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!content) error "KUBECONFIG_CONTENT missing from infra.env"
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: content
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                    sh "kubectl create namespace databases --dry-run=client -o yaml | kubectl apply -f -"
                }
            }
        }

        // ── INSTALL ──────────────────────────────────────────────────────────

        stage('ClickHouse') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm uninstall clickhouse -n databases --ignore-not-found || true
                        kubectl delete pvc -l app.kubernetes.io/instance=clickhouse -n databases --ignore-not-found || true
                        helm upgrade --install clickhouse helm/infra/databases/clickhouse \
                            --namespace databases \
                            --set image.tag=24.8-alpine \
                            --set persistentVolumeClaim.size=20Gi \
                            --set resources.requests.memory=1Gi \
                            --set resources.requests.cpu=500m \
                            --wait --timeout=8m
                        echo "Applying ClickHouse schemas..."
                        kubectl apply -f databases/clickhouse/ -n databases 2>/dev/null || true
                        echo "ClickHouse installed"
                    """
                }
            }
        }

        stage('Weaviate') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install weaviate helm/infra/databases/weaviate \
                            --namespace databases \
                            --set initContainers.sysctlInitContainer.enabled=false \
                            --set persistence.size=10Gi \
                            --set resources.requests.memory=512Mi \
                            --set resources.requests.cpu=250m \
                            --wait --timeout=8m
                        echo "Applying Weaviate schemas..."
                        kubectl apply -f databases/weaviate/ -n databases 2>/dev/null || true
                        echo "Weaviate installed"
                    """
                }
            }
        }

        stage('Neo4j') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install neo4j helm/infra/databases/neo4j \
                            --namespace databases \
                            --set neo4j.name=shopos-neo4j \
                            --set volumes.data.mode=defaultStorageClass \
                            --set neo4j.password=shopos-neo4j-password \
                            --wait --timeout=8m
                        echo "Applying Neo4j graph schemas..."
                        kubectl apply -f databases/neo4j/ -n databases 2>/dev/null || true
                        echo "Neo4j installed"
                    """
                }
            }
        }

        stage('Temporal') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        kubectl create namespace temporal-system --dry-run=client -o yaml | kubectl apply -f -
                        helm uninstall temporal -n temporal-system --ignore-not-found || true
                        kubectl delete pvc --all -n temporal-system --ignore-not-found || true
                        helm upgrade --install temporal helm/infra/databases/temporal \
                            --namespace temporal-system \
                            --set server.replicaCount=1 \
                            --set cassandra.config.cluster_size=1 \
                            --set elasticsearch.enabled=false \
                            --set grafana.enabled=false \
                            --set prometheus.enabled=false \
                            --timeout=3m
                        kubectl apply -f workflow/temporal/ -n temporal-system 2>/dev/null || true
                        echo "Temporal installed"
                    """
                }
            }
        }

        stage('Memcached') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install memcached helm/infra/databases/memcached \
                            --namespace databases \
                            --set replicaCount=3 \
                            --set resources.requests.memory=256Mi \
                            --set resources.requests.cpu=100m \
                            --set maxMemoryMb=256 \
                            --wait --timeout=5m
                        echo "Memcached installed"
                    """
                }
            }
        }

        stage('TimescaleDB') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install timescaledb helm/infra/databases/timescaledb \
                            --namespace databases \
                            --set image.tag=2.15-pg16-alpine \
                            --set persistence.size=20Gi \
                            --set resources.requests.memory=512Mi \
                            --set resources.requests.cpu=250m \
                            --wait --timeout=8m
                        echo "Applying TimescaleDB schemas..."
                        kubectl apply -f databases/timescaledb/ -n databases 2>/dev/null || true
                        echo "TimescaleDB installed"
                    """
                }
            }
        }

        stage('Apply OpenSearch Schemas') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                sh """
                    echo "Applying OpenSearch index templates..."
                    kubectl apply -f databases/opensearch/ -n monitoring 2>/dev/null || true
                    echo "OpenSearch schemas applied"
                """
            }
        }

        stage('Verify Databases') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                sh """
                    echo "=== Database pods in 'databases' namespace ==="
                    kubectl get pods -n databases
                    echo "=== Database services ==="
                    kubectl get svc -n databases
                    echo "=== Temporal ==="
                    kubectl get pods -n temporal-system 2>/dev/null || true
                    echo "=== Installed Helm releases ==="
                    helm list -n databases
                """
            }
        }

        // ── UNINSTALL ────────────────────────────────────────────────────────

        stage('Uninstall Temporal') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    helm uninstall temporal -n temporal-system --ignore-not-found || true
                    kubectl delete pvc --all -n temporal-system --ignore-not-found || true
                    kubectl delete namespace temporal-system --ignore-not-found || true
                """
            }
        }

        stage('Uninstall Neo4j') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    helm uninstall neo4j -n databases --ignore-not-found || true
                    kubectl delete pvc -l app.kubernetes.io/instance=neo4j -n databases --ignore-not-found || true
                """
            }
        }

        stage('Uninstall Weaviate') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    helm uninstall weaviate -n databases --ignore-not-found || true
                    kubectl delete pvc -l app.kubernetes.io/instance=weaviate -n databases --ignore-not-found || true
                """
            }
        }

        stage('Uninstall ClickHouse') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    helm uninstall clickhouse -n databases --ignore-not-found || true
                    kubectl delete pvc -l app.kubernetes.io/instance=clickhouse -n databases --ignore-not-found || true
                """
            }
        }

        stage('Uninstall Memcached') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    helm uninstall memcached -n databases --ignore-not-found || true
                """
            }
        }

        stage('Uninstall TimescaleDB') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    helm uninstall timescaledb -n databases --ignore-not-found || true
                    kubectl delete pvc -l app.kubernetes.io/instance=timescaledb -n databases --ignore-not-found || true
                """
            }
        }

        stage('Cleanup Namespaces') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh """
                    for ns in databases temporal-system cert-manager scylla-operator; do
                        kubectl delete all --all -n \$ns --ignore-not-found || true
                        kubectl delete pvc --all -n \$ns --ignore-not-found || true
                        kubectl delete configmap --all -n \$ns --ignore-not-found || true
                        kubectl delete secret --all -n \$ns --ignore-not-found || true
                        kubectl delete namespace \$ns --ignore-not-found || true
                    done
                    kubectl delete pv --all --ignore-not-found || true
                    kubectl delete clusterrolebinding -l app.kubernetes.io/instance=neo4j --ignore-not-found || true
                    kubectl delete clusterrolebinding -l app.kubernetes.io/instance=temporal --ignore-not-found || true
                    echo "Cleanup complete"
                """
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
        success { echo "${params.ACTION} of specialist databases completed successfully." }
        failure { echo "${params.ACTION} failed — check stage logs above." }
    }
}
