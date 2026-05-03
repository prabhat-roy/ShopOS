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
            description: 'INSTALL — deploy selected observability tools. UNINSTALL — remove selected.'
        )
        booleanParam(name: 'PROMETHEUS',           defaultValue: true,  description: 'Prometheus — metrics collection and alerting')
        booleanParam(name: 'ALERTMANAGER',         defaultValue: true,  description: 'Alertmanager — alert routing and grouping')
        booleanParam(name: 'THANOS',               defaultValue: true,  description: 'Thanos — long-term Prometheus storage and global query')
        booleanParam(name: 'VICTORIA_METRICS',     defaultValue: true,  description: 'VictoriaMetrics — high-performance metrics storage alternative')
        booleanParam(name: 'PUSHGATEWAY',          defaultValue: true,  description: 'Pushgateway — metrics push endpoint for batch jobs')
        booleanParam(name: 'BLACKBOX_EXPORTER',    defaultValue: true,  description: 'Blackbox Exporter — endpoint probing (HTTP, TCP, ICMP)')
        booleanParam(name: 'KUBE_STATE_METRICS',   defaultValue: true,  description: 'Kube State Metrics — K8s object state metrics')
        booleanParam(name: 'NODE_EXPORTER',        defaultValue: true,  description: 'Node Exporter — host-level hardware and OS metrics')
        booleanParam(name: 'ELASTICSEARCH',        defaultValue: true,  description: 'Elasticsearch — full-text search and log analytics')
        booleanParam(name: 'OPENSEARCH',           defaultValue: true,  description: 'OpenSearch — log analytics alternative')
        booleanParam(name: 'LOKI',                 defaultValue: true,  description: 'Grafana Loki — log aggregation system')
        booleanParam(name: 'FLUENTD',              defaultValue: true,  description: 'Fluentd — log collection and routing')
        booleanParam(name: 'FLUENT_BIT',           defaultValue: true,  description: 'Fluent Bit — lightweight log shipper')
        booleanParam(name: 'LOGSTASH',             defaultValue: true,  description: 'Logstash — log processing pipeline')
        booleanParam(name: 'JAEGER',               defaultValue: true,  description: 'Jaeger — distributed tracing')
        booleanParam(name: 'TEMPO',                defaultValue: true,  description: 'Grafana Tempo — distributed tracing backend')
        booleanParam(name: 'ZIPKIN',               defaultValue: true,  description: 'Zipkin — lightweight distributed tracing')
        booleanParam(name: 'GRAFANA',              defaultValue: true,  description: 'Grafana — dashboards and visualization')
        booleanParam(name: 'KIBANA',               defaultValue: true,  description: 'Kibana — Elasticsearch dashboards')
        booleanParam(name: 'OPENSEARCH_DASHBOARDS',defaultValue: true,  description: 'OpenSearch Dashboards — OpenSearch visualization')
        booleanParam(name: 'OTEL_COLLECTOR',       defaultValue: true,  description: 'OTel Collector — OpenTelemetry pipeline')
        booleanParam(name: 'SENTRY',               defaultValue: true,  description: 'Sentry OSS — error tracking and performance monitoring')
        booleanParam(name: 'GLITCHTIP',            defaultValue: true,  description: 'GlitchTip — Sentry-compatible error tracking alternative')
        booleanParam(name: 'PYRRA',                defaultValue: true,  description: 'Pyrra — SLO management and error budget tracking')
        booleanParam(name: 'SLOTH',                defaultValue: true,  description: 'Sloth — SLO definition and Prometheus rule generation')
        booleanParam(name: 'UPTIME_KUMA',          defaultValue: true,  description: 'Uptime Kuma — real-time uptime monitoring and status pages')
        booleanParam(name: 'PYROSCOPE',            defaultValue: true,  description: 'Grafana Pyroscope — continuous profiling')
        booleanParam(name: 'ROBUSTA',              defaultValue: true,  description: 'Robusta — Kubernetes alerting and runbook automation')
        booleanParam(name: 'OPENCOST',             defaultValue: true,  description: 'OpenCost — Kubernetes cost monitoring')
        booleanParam(name: 'GOLDILOCKS',           defaultValue: true,  description: 'Goldilocks — resource request/limit recommendations')
        booleanParam(name: 'GRAFANA_MIMIR',        defaultValue: true,  description: 'Grafana Mimir — multi-tenant Prometheus long-term storage')
        booleanParam(name: 'VICTORIA_LOGS',        defaultValue: true,  description: 'VictoriaLogs — high-volume log storage (overflow tier complementing Loki)')
        booleanParam(name: 'SIGNOZ',               defaultValue: true,  description: 'SigNoz — full-stack OTel-native observability for analytics-ai domain')
        booleanParam(name: 'KIALI',                defaultValue: true,  description: 'Kiali — Istio service mesh observability UI')
        booleanParam(name: 'NETDATA',              defaultValue: true,  description: 'Netdata — real-time per-second host metrics')
        booleanParam(name: 'PERSES',               defaultValue: true,  description: 'Perses — GitOps-native dashboards-as-code')
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

        stage('Prometheus') {
            when { expression { params.ACTION == 'INSTALL' && params.PROMETHEUS } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-prometheus.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-prometheus.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('prometheus')
                }
            }
        }

        stage('Alertmanager') {
            when { expression { params.ACTION == 'INSTALL' && params.ALERTMANAGER } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-alertmanager.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-alertmanager.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('alertmanager')
                }
            }
        }

        stage('Thanos') {
            when { expression { params.ACTION == 'INSTALL' && params.THANOS } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-thanos.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-thanos.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('thanos')
                }
            }
        }

        stage('Victoria Metrics') {
            when { expression { params.ACTION == 'INSTALL' && params.VICTORIA_METRICS } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-victoria-metrics.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-victoria-metrics.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('victoria-metrics')
                }
            }
        }

        stage('Pushgateway') {
            when { expression { params.ACTION == 'INSTALL' && params.PUSHGATEWAY } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-pushgateway.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pushgateway')
                }
            }
        }

        stage('Blackbox Exporter') {
            when { expression { params.ACTION == 'INSTALL' && params.BLACKBOX_EXPORTER } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-blackbox-exporter.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('blackbox-exporter')
                }
            }
        }

        stage('Kube State Metrics') {
            when { expression { params.ACTION == 'INSTALL' && params.KUBE_STATE_METRICS } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-kube-state-metrics.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kube-state-metrics')
                }
            }
        }

        stage('Node Exporter') {
            when { expression { params.ACTION == 'INSTALL' && params.NODE_EXPORTER } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-node-exporter.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('node-exporter')
                }
            }
        }

        stage('Elasticsearch') {
            when { expression { params.ACTION == 'INSTALL' && params.ELASTICSEARCH } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-elasticsearch.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-elasticsearch.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('elasticsearch')
                }
            }
        }

        stage('OpenSearch') {
            when { expression { params.ACTION == 'INSTALL' && params.OPENSEARCH } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-opensearch.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-opensearch.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('opensearch')
                }
            }
        }

        stage('Loki') {
            when { expression { params.ACTION == 'INSTALL' && params.LOKI } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-loki.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-loki.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('loki')
                }
            }
        }

        stage('Fluentd') {
            when { expression { params.ACTION == 'INSTALL' && params.FLUENTD } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-fluentd.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-fluentd.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('fluentd')
                }
            }
        }

        stage('Fluent Bit') {
            when { expression { params.ACTION == 'INSTALL' && params.FLUENT_BIT } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-fluent-bit.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-fluent-bit.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('fluent-bit')
                }
            }
        }

        stage('Logstash') {
            when { expression { params.ACTION == 'INSTALL' && params.LOGSTASH } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-logstash.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-logstash.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('logstash')
                }
            }
        }

        stage('Jaeger') {
            when { expression { params.ACTION == 'INSTALL' && params.JAEGER } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-jaeger.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-jaeger.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('jaeger')
                }
            }
        }

        stage('Tempo') {
            when { expression { params.ACTION == 'INSTALL' && params.TEMPO } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-tempo.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-tempo.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('tempo')
                }
            }
        }

        stage('Zipkin') {
            when { expression { params.ACTION == 'INSTALL' && params.ZIPKIN } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-zipkin.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-zipkin.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zipkin')
                }
            }
        }

        stage('Grafana') {
            when { expression { params.ACTION == 'INSTALL' && params.GRAFANA } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-grafana.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-grafana.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('grafana')
                }
            }
        }

        stage('Kibana') {
            when { expression { params.ACTION == 'INSTALL' && params.KIBANA } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-kibana.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-kibana.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kibana')
                }
            }
        }

        stage('OpenSearch Dashboards') {
            when { expression { params.ACTION == 'INSTALL' && params.OPENSEARCH_DASHBOARDS } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-opensearch-dashboards.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-opensearch-dashboards.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('opensearch-dashboards')
                }
            }
        }

        stage('OTel Collector') {
            when { expression { params.ACTION == 'INSTALL' && params.OTEL_COLLECTOR } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-otel-collector.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-otel-collector.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('otel-collector')
                }
            }
        }

        stage('Sentry') {
            when { expression { params.ACTION == 'INSTALL' && params.SENTRY } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-sentry.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-sentry.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('sentry')
                }
            }
        }

        stage('GlitchTip') {
            when { expression { params.ACTION == 'INSTALL' && params.GLITCHTIP } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-glitchtip.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-glitchtip.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('glitchtip')
                }
            }
        }

        stage('Pyrra') {
            when { expression { params.ACTION == 'INSTALL' && params.PYRRA } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-pyrra.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-pyrra.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pyrra')
                }
            }
        }

        stage('Sloth') {
            when { expression { params.ACTION == 'INSTALL' && params.SLOTH } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-sloth.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-sloth.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('sloth')
                }
            }
        }

        stage('Uptime Kuma') {
            when { expression { params.ACTION == 'INSTALL' && params.UPTIME_KUMA } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-uptime-kuma.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-uptime-kuma.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('uptime-kuma')
                }
            }
        }

        stage('Pyroscope') {
            when { expression { params.ACTION == 'INSTALL' && params.PYROSCOPE } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-pyroscope.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-pyroscope.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pyroscope')
                }
            }
        }

        stage('Robusta') {
            when { expression { params.ACTION == 'INSTALL' && params.ROBUSTA } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-robusta.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-robusta.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('monitoring')
                }
            }
        }

        stage('OpenCost') {
            when { expression { params.ACTION == 'INSTALL' && params.OPENCOST } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-opencost.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-opencost.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('monitoring')
                }
            }
        }

        stage('Goldilocks') {
            when { expression { params.ACTION == 'INSTALL' && params.GOLDILOCKS } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-goldilocks.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-goldilocks.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('monitoring')
                }
            }
        }

        // ── Vendored helm/infra/observability charts (no groovy script needed) ──

        stage('Grafana Mimir') {
            when { expression { params.ACTION == 'INSTALL' && params.GRAFANA_MIMIR } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install grafana-mimir helm/infra/observability/grafana-mimir \
                            --namespace monitoring \
                            --set minio.enabled=true \
                            --set ingester.replicas=1 \
                            --set store_gateway.replicas=1 \
                            --set compactor.replicas=1 \
                            --set distributor.replicas=1 \
                            --set querier.replicas=1 \
                            --wait --timeout=10m
                        echo "Grafana Mimir installed (multi-tenant Prometheus storage — partitions metrics per team)"
                    '''
                }
            }
        }

        stage('VictoriaLogs') {
            when { expression { params.ACTION == 'INSTALL' && params.VICTORIA_LOGS } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install victoria-logs helm/infra/observability/victoria-logs \
                            --namespace monitoring \
                            --set persistentVolume.size=20Gi \
                            --set resources.requests.memory=512Mi \
                            --set resources.requests.cpu=250m \
                            --wait --timeout=8m
                        echo "VictoriaLogs installed (high-volume overflow log storage — analytics + CDN access logs)"
                    '''
                }
            }
        }

        stage('SigNoz') {
            when { expression { params.ACTION == 'INSTALL' && params.SIGNOZ } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        kubectl create namespace signoz --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install signoz helm/infra/observability/signoz \
                            --namespace signoz \
                            --set clickhouse.replicaCount=1 \
                            --set zookeeper.replicaCount=1 \
                            --set queryService.resources.requests.memory=256Mi \
                            --set queryService.resources.requests.cpu=100m \
                            --wait --timeout=12m
                        echo "SigNoz installed (full-stack OTel-native observability — analytics-ai domain audience)"
                    '''
                }
            }
        }

        stage('Kiali') {
            when { expression { params.ACTION == 'INSTALL' && params.KIALI } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install kiali helm/infra/observability/kiali \
                            --namespace istio-system \
                            --set auth.strategy=anonymous \
                            --set deployment.ingress.enabled=true \
                            --set external_services.prometheus.url=http://prometheus.monitoring:9090 \
                            --set external_services.tracing.url=http://jaeger-query.tracing:16686 \
                            --set external_services.grafana.url=http://grafana.monitoring:3000 \
                            --wait --timeout=8m
                        echo "Kiali installed (Istio service mesh topology, mTLS status, traffic — SRE audience)"
                    '''
                }
            }
        }

        stage('Netdata') {
            when { expression { params.ACTION == 'INSTALL' && params.NETDATA } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install netdata helm/infra/observability/netdata \
                            --namespace monitoring \
                            --set parent.enabled=true \
                            --set parent.replicaCount=1 \
                            --set child.enabled=true \
                            --set parent.persistence.size=4Gi \
                            --wait --timeout=8m
                        echo "Netdata installed (real-time per-second host metrics — DaemonSet on every node)"
                    '''
                }
            }
        }

        stage('Perses') {
            when { expression { params.ACTION == 'INSTALL' && params.PERSES } }
            steps {
                catchError(buildResult: 'UNSTABLE', stageResult: 'FAILURE') {
                    sh '''
                        kubectl create namespace monitoring --dry-run=client -o yaml | kubectl apply -f -
                        helm upgrade --install perses helm/infra/observability/perses \
                            --namespace monitoring \
                            --set persistence.size=2Gi \
                            --set resources.requests.memory=256Mi \
                            --set resources.requests.cpu=100m \
                            --wait --timeout=6m
                        echo "Perses installed (GitOps-native dashboards-as-code — complements Grafana)"
                    '''
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall Perses') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PERSES } }
            steps {
                sh '''
                    helm uninstall perses -n monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Netdata') {
            when { expression { params.ACTION == 'UNINSTALL' && params.NETDATA } }
            steps {
                sh '''
                    helm uninstall netdata -n monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kiali') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KIALI } }
            steps {
                sh '''
                    helm uninstall kiali -n istio-system --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall SigNoz') {
            when { expression { params.ACTION == 'UNINSTALL' && params.SIGNOZ } }
            steps {
                sh '''
                    helm uninstall signoz -n signoz --ignore-not-found || true
                    kubectl delete pvc --all -n signoz --ignore-not-found || true
                    kubectl delete namespace signoz --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall VictoriaLogs') {
            when { expression { params.ACTION == 'UNINSTALL' && params.VICTORIA_LOGS } }
            steps {
                sh '''
                    helm uninstall victoria-logs -n monitoring --ignore-not-found || true
                    kubectl delete pvc -l app.kubernetes.io/instance=victoria-logs -n monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Grafana Mimir') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GRAFANA_MIMIR } }
            steps {
                sh '''
                    helm uninstall grafana-mimir -n monitoring --ignore-not-found || true
                    kubectl delete pvc -l app.kubernetes.io/instance=grafana-mimir -n monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Goldilocks') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GOLDILOCKS } }
            steps {
                sh '''
                    helm uninstall goldilocks -n monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OpenCost') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPENCOST } }
            steps {
                sh '''
                    helm uninstall opencost -n monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Robusta') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ROBUSTA } }
            steps {
                sh '''
                    helm uninstall robusta -n monitoring --ignore-not-found || true
                    kubectl delete namespace monitoring --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pyroscope') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PYROSCOPE } }
            steps {
                sh '''
                    helm uninstall pyroscope -n pyroscope --ignore-not-found || true
                    kubectl delete namespace pyroscope --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Uptime Kuma') {
            when { expression { params.ACTION == 'UNINSTALL' && params.UPTIME_KUMA } }
            steps {
                sh '''
                    helm uninstall uptime-kuma -n uptime-kuma --ignore-not-found || true
                    kubectl delete namespace uptime-kuma --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Sloth') {
            when { expression { params.ACTION == 'UNINSTALL' && params.SLOTH } }
            steps {
                sh '''
                    helm uninstall sloth -n sloth --ignore-not-found || true
                    kubectl delete namespace sloth --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pyrra') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PYRRA } }
            steps {
                sh '''
                    helm uninstall pyrra -n pyrra --ignore-not-found || true
                    kubectl delete namespace pyrra --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall GlitchTip') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GLITCHTIP } }
            steps {
                sh '''
                    helm uninstall glitchtip -n glitchtip --ignore-not-found || true
                    kubectl delete namespace glitchtip --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Sentry') {
            when { expression { params.ACTION == 'UNINSTALL' && params.SENTRY } }
            steps {
                sh '''
                    helm uninstall sentry -n sentry --ignore-not-found || true
                    kubectl delete namespace sentry --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OTel Collector') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OTEL_COLLECTOR } }
            steps {
                sh '''
                    helm uninstall otel-collector -n otel-collector --ignore-not-found || true
                    kubectl delete namespace otel-collector --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OpenSearch Dashboards') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPENSEARCH_DASHBOARDS } }
            steps {
                sh '''
                    helm uninstall opensearch-dashboards -n opensearch-dashboards --ignore-not-found || true
                    kubectl delete namespace opensearch-dashboards --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kibana') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KIBANA } }
            steps {
                sh '''
                    helm uninstall kibana -n kibana --ignore-not-found || true
                    kubectl delete namespace kibana --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Grafana') {
            when { expression { params.ACTION == 'UNINSTALL' && params.GRAFANA } }
            steps {
                sh '''
                    helm uninstall grafana -n grafana --ignore-not-found || true
                    kubectl delete pvc --all -n grafana --ignore-not-found || true
                    kubectl delete namespace grafana --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Zipkin') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ZIPKIN } }
            steps {
                sh '''
                    helm uninstall zipkin -n zipkin --ignore-not-found || true
                    kubectl delete namespace zipkin --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Tempo') {
            when { expression { params.ACTION == 'UNINSTALL' && params.TEMPO } }
            steps {
                sh '''
                    helm uninstall tempo -n tempo --ignore-not-found || true
                    kubectl delete namespace tempo --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Jaeger') {
            when { expression { params.ACTION == 'UNINSTALL' && params.JAEGER } }
            steps {
                sh '''
                    helm uninstall jaeger -n jaeger --ignore-not-found || true
                    kubectl delete namespace jaeger --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Logstash') {
            when { expression { params.ACTION == 'UNINSTALL' && params.LOGSTASH } }
            steps {
                sh '''
                    helm uninstall logstash -n logstash --ignore-not-found || true
                    kubectl delete namespace logstash --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Fluent Bit') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FLUENT_BIT } }
            steps {
                sh '''
                    helm uninstall fluent-bit -n fluent-bit --ignore-not-found || true
                    kubectl delete namespace fluent-bit --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Fluentd') {
            when { expression { params.ACTION == 'UNINSTALL' && params.FLUENTD } }
            steps {
                sh '''
                    helm uninstall fluentd -n fluentd --ignore-not-found || true
                    kubectl delete namespace fluentd --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Loki') {
            when { expression { params.ACTION == 'UNINSTALL' && params.LOKI } }
            steps {
                sh '''
                    helm uninstall loki -n loki --ignore-not-found || true
                    kubectl delete pvc --all -n loki --ignore-not-found || true
                    kubectl delete namespace loki --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall OpenSearch') {
            when { expression { params.ACTION == 'UNINSTALL' && params.OPENSEARCH } }
            steps {
                sh '''
                    helm uninstall opensearch -n opensearch --ignore-not-found || true
                    kubectl delete pvc --all -n opensearch --ignore-not-found || true
                    kubectl delete namespace opensearch --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Elasticsearch') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ELASTICSEARCH } }
            steps {
                sh '''
                    helm uninstall elasticsearch -n elasticsearch --ignore-not-found || true
                    kubectl delete pvc --all -n elasticsearch --ignore-not-found || true
                    kubectl delete namespace elasticsearch --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Node Exporter') {
            when { expression { params.ACTION == 'UNINSTALL' && params.NODE_EXPORTER } }
            steps {
                sh '''
                    helm uninstall node-exporter -n node-exporter --ignore-not-found || true
                    kubectl delete namespace node-exporter --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Kube State Metrics') {
            when { expression { params.ACTION == 'UNINSTALL' && params.KUBE_STATE_METRICS } }
            steps {
                sh '''
                    helm uninstall kube-state-metrics -n kube-state-metrics --ignore-not-found || true
                    kubectl delete namespace kube-state-metrics --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Blackbox Exporter') {
            when { expression { params.ACTION == 'UNINSTALL' && params.BLACKBOX_EXPORTER } }
            steps {
                sh '''
                    helm uninstall blackbox-exporter -n blackbox-exporter --ignore-not-found || true
                    kubectl delete namespace blackbox-exporter --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Pushgateway') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PUSHGATEWAY } }
            steps {
                sh '''
                    helm uninstall pushgateway -n pushgateway --ignore-not-found || true
                    kubectl delete namespace pushgateway --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Victoria Metrics') {
            when { expression { params.ACTION == 'UNINSTALL' && params.VICTORIA_METRICS } }
            steps {
                sh '''
                    helm uninstall victoria-metrics -n victoria-metrics --ignore-not-found || true
                    kubectl delete pvc --all -n victoria-metrics --ignore-not-found || true
                    kubectl delete namespace victoria-metrics --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Thanos') {
            when { expression { params.ACTION == 'UNINSTALL' && params.THANOS } }
            steps {
                sh '''
                    helm uninstall thanos -n thanos --ignore-not-found || true
                    kubectl delete pvc --all -n thanos --ignore-not-found || true
                    kubectl delete namespace thanos --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Alertmanager') {
            when { expression { params.ACTION == 'UNINSTALL' && params.ALERTMANAGER } }
            steps {
                sh '''
                    helm uninstall alertmanager -n alertmanager --ignore-not-found || true
                    kubectl delete pvc --all -n alertmanager --ignore-not-found || true
                    kubectl delete namespace alertmanager --ignore-not-found || true
                '''
            }
        }

        stage('Uninstall Prometheus') {
            when { expression { params.ACTION == 'UNINSTALL' && params.PROMETHEUS } }
            steps {
                sh '''
                    helm uninstall prometheus -n prometheus --ignore-not-found || true
                    kubectl delete pvc --all -n prometheus --ignore-not-found || true
                    kubectl delete namespace prometheus --ignore-not-found || true
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
