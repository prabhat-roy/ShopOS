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
            description: 'INSTALL — deploy developer tooling stack. UNINSTALL — remove all.'
        )
        booleanParam(name: 'PGADMIN',            defaultValue: true, description: 'pgAdmin 4 — PostgreSQL management UI')
        booleanParam(name: 'MONGO_EXPRESS',       defaultValue: true, description: 'Mongo Express — MongoDB management UI')
        booleanParam(name: 'REDIS_COMMANDER',     defaultValue: true, description: 'Redis Commander — Redis key browser')
        booleanParam(name: 'BYTEBASE',            defaultValue: true, description: 'Bytebase — schema change management with approval workflows')
        booleanParam(name: 'SUPERSET',            defaultValue: true, description: 'Apache Superset — BI and data visualization on ClickHouse/Postgres')
        booleanParam(name: 'MARQUEZ',             defaultValue: true, description: 'Marquez — OpenLineage data lineage server')
        booleanParam(name: 'GREAT_EXPECTATIONS',  defaultValue: true, description: 'Great Expectations — data quality assertions')
        booleanParam(name: 'APACHE_ATLAS',        defaultValue: true, description: 'Apache Atlas — data catalog and governance')
        booleanParam(name: 'PACT_BROKER',         defaultValue: true, description: 'Pact Broker — consumer-driven contract test registry')
        booleanParam(name: 'BOTKUBE',             defaultValue: true, description: 'Botkube — Kubernetes event alerts to Slack')
        booleanParam(name: 'K8SGPT',              defaultValue: true, description: 'k8sGPT operator — AI-powered Kubernetes diagnostics')
        booleanParam(name: 'OPENCOST',            defaultValue: true, description: 'OpenCost — per-namespace Kubernetes cost attribution')
        booleanParam(name: 'TELEPORT',            defaultValue: true, description: 'Teleport — zero-trust SSH and Kubernetes access')
        booleanParam(name: 'CONDUKTOR_GATEWAY',   defaultValue: true, description: 'Conduktor Gateway — Kafka policy enforcement proxy')
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
                    sh "kubectl create namespace tooling --dry-run=client -o yaml | kubectl apply -f -"
                    sh "kubectl create namespace data-platform --dry-run=client -o yaml | kubectl apply -f -"
                    sh "kubectl create namespace contract-testing --dry-run=client -o yaml | kubectl apply -f -"
                }
            }
        }

        // ── Database Management UIs ───────────────────────────────────────────

        stage('pgAdmin') {
            when { expression { params.ACTION == 'INSTALL' && params.PGADMIN } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install pgadmin charts/db-management/pgadmin \
                            --namespace tooling \
                            --set env.email=admin@shopos.dev \
                            --set env.password=admin \
                            --set persistence.size=1Gi \
                            --wait --timeout=5m
                        echo "pgAdmin installed — http://pgadmin.tooling.svc.cluster.local"
                    """
                }
            }
        }

        stage('Mongo Express') {
            when { expression { params.ACTION == 'INSTALL' && params.MONGO_EXPRESS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install mongo-express charts/db-management/mongo-express \
                            --namespace tooling \
                            --set env.mongodbServer=mongodb.default.svc.cluster.local \
                            --wait --timeout=5m
                        echo "Mongo Express installed"
                    """
                }
            }
        }

        stage('Redis Commander') {
            when { expression { params.ACTION == 'INSTALL' && params.REDIS_COMMANDER } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install redis-commander charts/db-management/redis-commander \
                            --namespace tooling \
                            --set env.redisHosts=local:redis.default.svc.cluster.local:6379 \
                            --wait --timeout=5m
                        echo "Redis Commander installed"
                    """
                }
            }
        }

        stage('Bytebase') {
            when { expression { params.ACTION == 'INSTALL' && params.BYTEBASE } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install bytebase charts/db-management/bytebase \
                            --namespace tooling \
                            --set env.pgUrl=postgresql://postgres:postgres@postgres.default.svc.cluster.local/bytebase?sslmode=disable \
                            --set persistence.size=5Gi \
                            --wait --timeout=8m
                        echo "Bytebase installed — schema change management ready"
                    """
                }
            }
        }

        // ── Analytics & Data Platform ─────────────────────────────────────────

        stage('Apache Superset') {
            when { expression { params.ACTION == 'INSTALL' && params.SUPERSET } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install superset charts/analytics/superset \
                            --namespace data-platform \
                            --set env.secretKey=shopos-superset-secret \
                            --set env.databaseUrl=postgresql+psycopg2://postgres:postgres@postgres.default.svc.cluster.local/superset \
                            --set env.redisUrl=redis://redis.default.svc.cluster.local:6379/1 \
                            --set persistence.size=5Gi \
                            --timeout=10m
                        echo "Superset installed — BI dashboards on ClickHouse + Postgres"
                    """
                }
            }
        }

        stage('Marquez') {
            when { expression { params.ACTION == 'INSTALL' && params.MARQUEZ } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install marquez charts/data-lineage/marquez \
                            --namespace data-platform \
                            --set env.postgresHost=postgres.default.svc.cluster.local \
                            --set env.postgresUser=postgres \
                            --set env.postgresPassword=postgres \
                            --set env.postgresDb=marquez \
                            --wait --timeout=5m
                        helm upgrade --install marquez-web charts/data-lineage/marquez-web \
                            --namespace data-platform \
                            --set env.marquezHost=marquez.data-platform.svc.cluster.local \
                            --set env.marquezPort=5000 \
                            --wait --timeout=3m
                        kubectl apply -f data-lineage/openlineage/ -n data-platform 2>/dev/null || true
                        echo "Marquez + OpenLineage installed — data lineage tracking ready"
                    """
                }
            }
        }

        stage('Great Expectations') {
            when { expression { params.ACTION == 'INSTALL' && params.GREAT_EXPECTATIONS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install great-expectations charts/data-quality/great-expectations \
                            --namespace data-platform \
                            --set persistence.size=2Gi \
                            --wait --timeout=5m
                        kubectl create configmap ge-config \
                            --from-file=data-quality/great-expectations/ \
                            -n data-platform --dry-run=client -o yaml | kubectl apply -f -
                        echo "Great Expectations installed — data quality docs at :4000"
                    """
                }
            }
        }

        stage('Apache Atlas') {
            when { expression { params.ACTION == 'INSTALL' && params.APACHE_ATLAS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install apache-atlas charts/data-catalog/apache-atlas \
                            --namespace data-platform \
                            --set resources.requests.memory=1Gi \
                            --set resources.limits.memory=2Gi \
                            --set persistence.size=10Gi \
                            --timeout=15m
                        echo "Apache Atlas installed — data catalog at :21000"
                    """
                }
            }
        }

        // ── Contract Testing ──────────────────────────────────────────────────

        stage('Pact Broker') {
            when { expression { params.ACTION == 'INSTALL' && params.PACT_BROKER } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install pact-broker charts/testing/pact-broker \
                            --namespace contract-testing \
                            --set env.databaseUrl=postgres://postgres:postgres@postgres.default.svc.cluster.local/pact_broker \
                            --set env.basicAuthUsername=admin \
                            --set env.basicAuthPassword=admin \
                            --wait --timeout=5m
                        echo "Applying example Pact contracts..."
                        kubectl create configmap pact-contracts \
                            --from-file=testing/pact/consumer/ \
                            -n contract-testing --dry-run=client -o yaml | kubectl apply -f -
                        echo "Pact Broker installed — contract registry at :9292"
                    """
                }
            }
        }

        // ── Kubernetes Operators ──────────────────────────────────────────────

        stage('Botkube') {
            when { expression { params.ACTION == 'INSTALL' && params.BOTKUBE } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add botkube https://charts.botkube.io --force-update || true
                        helm repo update botkube || true
                        helm upgrade --install botkube botkube/botkube \
                            --namespace botkube --create-namespace \
                            --values kubernetes/botkube/botkube-values.yaml \
                            --wait --timeout=5m
                        echo "Botkube installed — K8s event alerts to Slack"
                    """
                }
            }
        }

        stage('k8sGPT Operator') {
            when { expression { params.ACTION == 'INSTALL' && params.K8SGPT } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add k8sgpt https://charts.k8sgpt.ai --force-update || true
                        helm repo update k8sgpt || true
                        helm upgrade --install k8sgpt-operator k8sgpt/k8sgpt-operator \
                            --namespace k8sgpt-operator-system --create-namespace \
                            --values kubernetes/k8sgpt/k8sgpt-operator-values.yaml \
                            --wait --timeout=5m
                        kubectl apply -f kubernetes/k8sgpt/k8sgpt-cr.yaml || true
                        echo "k8sGPT operator installed — Kubernetes diagnostics active"
                    """
                }
            }
        }

        stage('OpenCost') {
            when { expression { params.ACTION == 'INSTALL' && params.OPENCOST } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add opencost https://opencost.github.io/opencost-helm-chart --force-update || true
                        helm repo update opencost || true
                        kubectl apply -f kubernetes/opencost/opencost-serviceaccount.yaml || true
                        helm upgrade --install opencost opencost/opencost \
                            --namespace monitoring --create-namespace \
                            --values kubernetes/opencost/opencost-values.yaml \
                            --wait --timeout=5m
                        echo "OpenCost installed — cost attribution by namespace/service"
                    """
                }
            }
        }

        // ── Access Control ────────────────────────────────────────────────────

        stage('Teleport') {
            when { expression { params.ACTION == 'INSTALL' && params.TELEPORT } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm repo add teleport https://charts.releases.teleport.dev --force-update || true
                        helm repo update teleport || true
                        helm upgrade --install teleport teleport/teleport-cluster \
                            --namespace teleport-cluster --create-namespace \
                            --values security/teleport/teleport-values.yaml \
                            --wait --timeout=10m
                        kubectl apply -f security/teleport/roles/ -n teleport-cluster || true
                        echo "Teleport installed — zero-trust SSH + K8s access at shopos.internal"
                    """
                }
            }
        }

        // ── Messaging Governance ──────────────────────────────────────────────

        stage('Conduktor Gateway') {
            when { expression { params.ACTION == 'INSTALL' && params.CONDUKTOR_GATEWAY } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh """
                        helm upgrade --install conduktor-gateway charts/messaging/conduktor-gateway \
                            --namespace messaging --create-namespace \
                            --set env.kafkaBootstrapServers=kafka.default.svc.cluster.local:9092 \
                            --wait --timeout=5m
                        kubectl create configmap conduktor-interceptors \
                            --from-file=messaging/conduktor/gateway-config.yaml \
                            -n messaging --dry-run=client -o yaml | kubectl apply -f -
                        echo "Conduktor Gateway installed — Kafka policy enforcement active"
                    """
                }
            }
        }

        // ── Verify ────────────────────────────────────────────────────────────

        stage('Verify Tooling') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                sh """
                    echo "=== Tooling namespace ==="
                    kubectl get pods -n tooling 2>/dev/null || true
                    echo "=== Data platform namespace ==="
                    kubectl get pods -n data-platform 2>/dev/null || true
                    echo "=== Contract testing namespace ==="
                    kubectl get pods -n contract-testing 2>/dev/null || true
                    echo "=== Botkube ==="
                    kubectl get pods -n botkube 2>/dev/null || true
                    echo "=== k8sGPT operator ==="
                    kubectl get pods -n k8sgpt-operator-system 2>/dev/null || true
                    echo "=== OpenCost ==="
                    kubectl get pods -n monitoring -l app.kubernetes.io/name=opencost 2>/dev/null || true
                    echo "=== Teleport ==="
                    kubectl get pods -n teleport-cluster 2>/dev/null || true
                    echo "=== Conduktor Gateway ==="
                    kubectl get pods -n messaging 2>/dev/null || true
                """
            }
        }

        // ── UNINSTALL ─────────────────────────────────────────────────────────

        stage('Uninstall Conduktor Gateway') {
            when { expression { params.ACTION == 'UNINSTALL' && params.CONDUKTOR_GATEWAY } }
            steps { sh 'helm uninstall conduktor-gateway -n messaging --ignore-not-found || true' }
        }

        stage('Uninstall Teleport') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TELEPORT } }
            steps {
                sh '''
                    helm uninstall teleport -n teleport-cluster --ignore-not-found || true
                    kubectl delete pvc --all -n teleport-cluster --ignore-not-found || true
                    kubectl delete namespace teleport-cluster --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OpenCost') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPENCOST } }
            steps { sh 'helm uninstall opencost -n monitoring --ignore-not-found || true' }
        }

        stage('Uninstall k8sGPT') {
            when { expression { params.ACTION == 'UNINSTALL' && params.K8SGPT } }
            steps {
                sh '''
                    kubectl delete -f kubernetes/k8sgpt/k8sgpt-cr.yaml --ignore-not-found || true
                    helm uninstall k8sgpt-operator -n k8sgpt-operator-system --ignore-not-found || true
                    kubectl delete namespace k8sgpt-operator-system --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Botkube') {
            when { expression { params.ACTION == 'UNINSTALL' && params.BOTKUBE } }
            steps {
                sh '''
                    helm uninstall botkube -n botkube --ignore-not-found || true
                    kubectl delete namespace botkube --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pact Broker') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PACT_BROKER } }
            steps {
                sh '''
                    helm uninstall pact-broker -n contract-testing --ignore-not-found || true
                    kubectl delete namespace contract-testing --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Data Platform') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    for svc in apache-atlas great-expectations marquez marquez-web superset; do
                        helm uninstall $svc -n data-platform --ignore-not-found || true
                    done
                    kubectl delete pvc --all -n data-platform --ignore-not-found || true
                    kubectl delete namespace data-platform --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Tooling UIs') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                sh '''
                    for svc in bytebase redis-commander mongo-express pgadmin; do
                        helm uninstall $svc -n tooling --ignore-not-found || true
                    done
                    kubectl delete pvc --all -n tooling --ignore-not-found || true
                    kubectl delete namespace tooling --ignore-not-found || true
                '''
            }
        }
    }

    post {
        always {
            sh 'test -f infra.env && cp infra.env /var/lib/jenkins/infra.env || true'
            sh "rm -f ${env.WORKSPACE}/kubeconfig 2>/dev/null || true"
        }
        success { echo "${params.ACTION} of developer tooling completed successfully." }
        failure { echo "${params.ACTION} failed — check stage logs above." }
    }
}
