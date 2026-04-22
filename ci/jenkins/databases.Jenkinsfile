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
            description: 'INSTALL — deploy specialist databases. UNINSTALL — remove all.'
        )
        booleanParam(name: 'CLICKHOUSE',  defaultValue: true,  description: 'ClickHouse OLAP database')
        booleanParam(name: 'WEAVIATE',    defaultValue: true,  description: 'Weaviate vector database')
        booleanParam(name: 'NEO4J',       defaultValue: true,  description: 'Neo4j graph database')
        booleanParam(name: 'SCYLLADB',    defaultValue: true,  description: 'ScyllaDB high-throughput database')
        booleanParam(name: 'TEMPORAL',    defaultValue: true,  description: 'Temporal workflow engine')
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
            when { allOf { expression { params.ACTION == 'INSTALL' }; expression { params.CLICKHOUSE } } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add pascaliske https://charts.pascaliske.dev || true
                        helm repo update
                        helm uninstall clickhouse -n databases --ignore-not-found || true
                        kubectl delete pvc -l app.kubernetes.io/instance=clickhouse -n databases --ignore-not-found || true
                        helm upgrade --install clickhouse pascaliske/clickhouse \
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
            when { allOf { expression { params.ACTION == 'INSTALL' }; expression { params.WEAVIATE } } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add weaviate https://weaviate.github.io/weaviate-helm || true
                        helm repo update
                        helm upgrade --install weaviate weaviate/weaviate \
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
            when { allOf { expression { params.ACTION == 'INSTALL' }; expression { params.NEO4J } } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add neo4j https://helm.neo4j.com/neo4j || true
                        helm repo update
                        helm upgrade --install neo4j neo4j/neo4j \
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

        stage('ScyllaDB') {
            when { allOf { expression { params.ACTION == 'INSTALL' }; expression { params.SCYLLADB } } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add scylla https://scylla-operator-charts.storage.googleapis.com/stable || true
                        helm repo add jetstack https://charts.jetstack.io || true
                        helm repo update
                        helm uninstall cert-manager -n cert-manager --ignore-not-found || true
                        kubectl delete namespace cert-manager --ignore-not-found || true
                        helm upgrade --install cert-manager jetstack/cert-manager \
                            --namespace cert-manager --create-namespace \
                            --set installCRDs=true \
                            --set startupapicheck.enabled=false \
                            --wait --timeout=5m
                        kubectl rollout status deployment/cert-manager-webhook -n cert-manager --timeout=180s
                        kubectl rollout status deployment/cert-manager-cainjector -n cert-manager --timeout=180s
                        echo "Waiting for cert-manager CA bundle injection (up to 3m)..."
                        for i in \$(seq 1 18); do
                            CA=\$(kubectl get validatingwebhookconfiguration cert-manager-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}' 2>/dev/null || echo "")
                            if [ -n "\$CA" ]; then echo "CA bundle injected after \$i attempts"; break; fi
                            echo "Waiting (\$i/18)..."; sleep 10
                        done
                        CA=\$(kubectl get validatingwebhookconfiguration cert-manager-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}' 2>/dev/null || echo "")
                        if [ -z "\$CA" ]; then
                            echo "CA bundle not injected — removing cert-manager webhook to unblock scylla-operator..."
                            kubectl delete validatingwebhookconfiguration cert-manager-webhook --ignore-not-found || true
                            kubectl delete mutatingwebhookconfiguration cert-manager-webhook --ignore-not-found || true
                        fi
                        helm upgrade --install scylla-operator scylla/scylla-operator \
                            --namespace scylla-operator --create-namespace \
                            --wait --timeout=5m
                        helm upgrade --install scylla scylla/scylla \
                            --namespace databases \
                            --set developerMode=true \
                            --wait --timeout=10m
                        echo "Applying ScyllaDB keyspace schemas..."
                        kubectl apply -f databases/scylladb/ -n databases 2>/dev/null || true
                        echo "ScyllaDB installed"
                    """
                }
            }
        }

        stage('Temporal') {
            when { allOf { expression { params.ACTION == 'INSTALL' }; expression { params.TEMPORAL } } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add temporal https://go.temporal.io/helm-charts || true
                        helm repo update
                        kubectl create namespace temporal-system --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install temporal temporal/temporal \
                            --version 0.74.0 \
                            --namespace temporal-system \
                            --set server.replicaCount=1 \
                            --set cassandra.config.cluster_size=1 \
                            --set elasticsearch.enabled=false \
                            --set grafana.enabled=false \
                            --set prometheus.enabled=false \
                            --set schema.setup.enabled=false \
                            --set schema.update.enabled=false
                        kubectl apply -f workflow/temporal/ -n temporal-system 2>/dev/null || true
                        echo "Temporal installed"
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
                """
            }
        }

        // ── UNINSTALL ────────────────────────────────────────────────────────

        stage('Uninstall Temporal') {
            when { allOf { expression { params.ACTION == 'UNINSTALL' }; expression { params.TEMPORAL } } }
            steps {
                sh "helm uninstall temporal -n temporal-system --ignore-not-found || true"
            }
        }

        stage('Uninstall ScyllaDB') {
            when { allOf { expression { params.ACTION == 'UNINSTALL' }; expression { params.SCYLLADB } } }
            steps {
                sh """
                    helm uninstall scylla -n databases --ignore-not-found || true
                    helm uninstall scylla-operator -n scylla-operator --ignore-not-found || true
                """
            }
        }

        stage('Uninstall Neo4j') {
            when { allOf { expression { params.ACTION == 'UNINSTALL' }; expression { params.NEO4J } } }
            steps {
                sh "helm uninstall neo4j -n databases --ignore-not-found || true"
            }
        }

        stage('Uninstall Weaviate') {
            when { allOf { expression { params.ACTION == 'UNINSTALL' }; expression { params.WEAVIATE } } }
            steps {
                sh "helm uninstall weaviate -n databases --ignore-not-found || true"
            }
        }

        stage('Uninstall ClickHouse') {
            when { allOf { expression { params.ACTION == 'UNINSTALL' }; expression { params.CLICKHOUSE } } }
            steps {
                sh "helm uninstall clickhouse -n databases --ignore-not-found || true"
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
