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
            description: 'INSTALL — deploy, configure and apply K8s enhancements for all observability tools. UNINSTALL — remove all.'
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
                }
            }
        }

        // ── INSTALL + CONFIGURE + K8s ENHANCEMENTS ───────────────────────────

        stage('Prometheus') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-prometheus.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-prometheus.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('prometheus')
                }
            }
        }

        stage('Alertmanager') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-alertmanager.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-alertmanager.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('alertmanager')
                }
            }
        }

        stage('Thanos') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-thanos.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-thanos.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('thanos')
                }
            }
        }

        stage('Victoria Metrics') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-victoria-metrics.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-victoria-metrics.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('victoria-metrics')
                }
            }
        }

        stage('Pushgateway') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-pushgateway.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pushgateway')
                }
            }
        }

        stage('Blackbox Exporter') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-blackbox-exporter.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('blackbox-exporter')
                }
            }
        }

        stage('Kube State Metrics') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-kube-state-metrics.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kube-state-metrics')
                }
            }
        }

        stage('Node Exporter') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-node-exporter.groovy'; s()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('node-exporter')
                }
            }
        }

        stage('Elasticsearch') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-elasticsearch.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-elasticsearch.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('elasticsearch')
                }
            }
        }

        stage('OpenSearch') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-opensearch.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-opensearch.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('opensearch')
                }
            }
        }

        stage('Loki') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-loki.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-loki.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('loki')
                }
            }
        }

        stage('Fluentd') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-fluentd.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-fluentd.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('fluentd')
                }
            }
        }

        stage('Fluent Bit') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-fluent-bit.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-fluent-bit.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('fluent-bit')
                }
            }
        }

        stage('Logstash') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-logstash.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-logstash.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('logstash')
                }
            }
        }

        stage('Jaeger') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-jaeger.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-jaeger.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('jaeger')
                }
            }
        }

        stage('Tempo') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-tempo.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-tempo.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('tempo')
                }
            }
        }

        stage('Zipkin') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-zipkin.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-zipkin.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('zipkin')
                }
            }
        }

        stage('Grafana') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-grafana.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-grafana.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('grafana')
                }
            }
        }

        stage('Kibana') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-kibana.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-kibana.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('kibana')
                }
            }
        }

        stage('OpenSearch Dashboards') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-opensearch-dashboards.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-opensearch-dashboards.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('opensearch-dashboards')
                }
            }
        }

        stage('OTel Collector') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-otel-collector.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-otel-collector.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('otel-collector')
                }
            }
        }

        stage('Sentry') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-sentry.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-sentry.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('sentry')
                }
            }
        }

        stage('GlitchTip') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-glitchtip.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-glitchtip.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('glitchtip')
                }
            }
        }

        stage('Pyrra') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-pyrra.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-pyrra.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pyrra')
                }
            }
        }

        stage('Sloth') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-sloth.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-sloth.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('sloth')
                }
            }
        }

        stage('Uptime Kuma') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-uptime-kuma.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-uptime-kuma.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('uptime-kuma')
                }
            }
        }

        stage('Pyroscope') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    def s = load 'scripts/groovy/observability-install-pyroscope.groovy'; s()
                    def c = load 'scripts/groovy/observability-configure-pyroscope.groovy'; c()
                    def e = load 'scripts/groovy/apply-k8s-enhancements.groovy'; e('pyroscope')
                }
            }
        }

        // ── UNINSTALL (reverse order) ─────────────────────────────────────────

        stage('Uninstall Pyroscope') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall pyroscope -n pyroscope --ignore-not-found || true' }
        }

        stage('Uninstall Uptime Kuma') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall uptime-kuma -n uptime-kuma --ignore-not-found || true' }
        }

        stage('Uninstall Sloth') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall sloth -n sloth --ignore-not-found || true' }
        }

        stage('Uninstall Pyrra') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall pyrra -n pyrra --ignore-not-found || true' }
        }

        stage('Uninstall GlitchTip') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall glitchtip -n glitchtip --ignore-not-found || true' }
        }

        stage('Uninstall Sentry') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall sentry -n sentry --ignore-not-found || true' }
        }

        stage('Uninstall OTel Collector') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall otel-collector -n otel-collector --ignore-not-found || true' }
        }

        stage('Uninstall OpenSearch Dashboards') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall opensearch-dashboards -n opensearch-dashboards --ignore-not-found || true' }
        }

        stage('Uninstall Kibana') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kibana -n kibana --ignore-not-found || true' }
        }

        stage('Uninstall Grafana') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall grafana -n grafana --ignore-not-found || true' }
        }

        stage('Uninstall Zipkin') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall zipkin -n zipkin --ignore-not-found || true' }
        }

        stage('Uninstall Tempo') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall tempo -n tempo --ignore-not-found || true' }
        }

        stage('Uninstall Jaeger') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall jaeger -n jaeger --ignore-not-found || true' }
        }

        stage('Uninstall Logstash') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall logstash -n logstash --ignore-not-found || true' }
        }

        stage('Uninstall Fluent Bit') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall fluent-bit -n fluent-bit --ignore-not-found || true' }
        }

        stage('Uninstall Fluentd') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall fluentd -n fluentd --ignore-not-found || true' }
        }

        stage('Uninstall Loki') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall loki -n loki --ignore-not-found || true' }
        }

        stage('Uninstall OpenSearch') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall opensearch -n opensearch --ignore-not-found || true' }
        }

        stage('Uninstall Elasticsearch') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall elasticsearch -n elasticsearch --ignore-not-found || true' }
        }

        stage('Uninstall Node Exporter') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall node-exporter -n node-exporter --ignore-not-found || true' }
        }

        stage('Uninstall Kube State Metrics') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall kube-state-metrics -n kube-state-metrics --ignore-not-found || true' }
        }

        stage('Uninstall Blackbox Exporter') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall blackbox-exporter -n blackbox-exporter --ignore-not-found || true' }
        }

        stage('Uninstall Pushgateway') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall pushgateway -n pushgateway --ignore-not-found || true' }
        }

        stage('Uninstall Victoria Metrics') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall victoria-metrics -n victoria-metrics --ignore-not-found || true' }
        }

        stage('Uninstall Thanos') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall thanos -n thanos --ignore-not-found || true' }
        }

        stage('Uninstall Alertmanager') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall alertmanager -n alertmanager --ignore-not-found || true' }
        }

        stage('Uninstall Prometheus') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps { sh 'helm uninstall prometheus -n prometheus --ignore-not-found || true' }
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
