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
            description: 'INSTALL/UNINSTALL deploys observability tools on Kubernetes. CONFIGURE applies post-install setup (datasources, dashboards, scrape configs, pipelines).'
        )

        // ── Metrics ───────────────────────────────────────────────────────
        booleanParam(name: 'PROMETHEUS',             defaultValue: false, description: 'Prometheus — open source monitoring and alerting toolkit (CNCF)')
        booleanParam(name: 'ALERTMANAGER',           defaultValue: false, description: 'Alertmanager — handles alerts from Prometheus, routes to receivers')
        booleanParam(name: 'THANOS',                 defaultValue: false, description: 'Thanos — highly available Prometheus with long-term storage')
        booleanParam(name: 'VICTORIA_METRICS',       defaultValue: false, description: 'VictoriaMetrics — fast cost-effective monitoring and time series DB')
        booleanParam(name: 'PUSHGATEWAY',            defaultValue: false, description: 'Prometheus Pushgateway — push metrics from batch and short-lived jobs')
        booleanParam(name: 'BLACKBOX_EXPORTER',      defaultValue: false, description: 'Blackbox Exporter — probes endpoints over HTTP, HTTPS, DNS, TCP, ICMP')
        booleanParam(name: 'KUBE_STATE_METRICS',     defaultValue: false, description: 'kube-state-metrics — generates metrics from Kubernetes object state')
        booleanParam(name: 'NODE_EXPORTER',          defaultValue: false, description: 'Node Exporter — exposes hardware and OS metrics from Linux hosts')

        // ── Logs ──────────────────────────────────────────────────────────
        booleanParam(name: 'ELASTICSEARCH',          defaultValue: false, description: 'Elasticsearch — distributed search and analytics engine (ELK stack backend)')
        booleanParam(name: 'OPENSEARCH',             defaultValue: false, description: 'OpenSearch — open source distributed search and log analytics engine')
        booleanParam(name: 'LOKI',                   defaultValue: false, description: 'Grafana Loki — horizontally scalable log aggregation system')
        booleanParam(name: 'FLUENTD',                defaultValue: false, description: 'Fluentd — unified open source data collector for logging layer')
        booleanParam(name: 'FLUENT_BIT',             defaultValue: false, description: 'Fluent Bit — lightweight high performance log processor and forwarder')
        booleanParam(name: 'LOGSTASH',               defaultValue: false, description: 'Logstash — dynamic data collection pipeline with pluggable input/output')

        // ── Tracing ───────────────────────────────────────────────────────
        booleanParam(name: 'JAEGER',                 defaultValue: false, description: 'Jaeger — end-to-end distributed tracing and root cause analysis (CNCF)')
        booleanParam(name: 'TEMPO',                  defaultValue: false, description: 'Grafana Tempo — high scale distributed tracing backend')
        booleanParam(name: 'ZIPKIN',                 defaultValue: false, description: 'Zipkin — distributed tracing system for latency troubleshooting')

        // ── Dashboards ────────────────────────────────────────────────────
        booleanParam(name: 'GRAFANA',                defaultValue: false, description: 'Grafana — open source analytics and interactive visualization platform')
        booleanParam(name: 'KIBANA',                 defaultValue: false, description: 'Kibana — data visualization and exploration UI for Elasticsearch')
        booleanParam(name: 'OPENSEARCH_DASHBOARDS',  defaultValue: false, description: 'OpenSearch Dashboards — visualization UI for OpenSearch')

        // ── Instrumentation ───────────────────────────────────────────────
        booleanParam(name: 'OTEL_COLLECTOR',         defaultValue: false, description: 'OpenTelemetry Collector — vendor-agnostic telemetry collection pipeline')

        // ── Error Tracking ────────────────────────────────────────────────
        booleanParam(name: 'SENTRY',                 defaultValue: false, description: 'Sentry OSS — open source error tracking and performance monitoring')
        booleanParam(name: 'GLITCHTIP',              defaultValue: false, description: 'GlitchTip — open source Sentry-compatible error tracking platform')

        // ── SLO ───────────────────────────────────────────────────────────
        booleanParam(name: 'PYRRA',                  defaultValue: false, description: 'Pyrra — SLO management and alerting UI for Prometheus')
        booleanParam(name: 'SLOTH',                  defaultValue: false, description: 'Sloth — SLO generator for Prometheus using simple declarative specs')
        booleanParam(name: 'UPTIME_KUMA',            defaultValue: false, description: 'Uptime Kuma — self-hosted uptime monitoring and status pages')

        // ── Profiling ─────────────────────────────────────────────────────
        booleanParam(name: 'PYROSCOPE',              defaultValue: false, description: 'Grafana Pyroscope — continuous profiling for all applications')
    }

    stages {
        stage('Git Fetch') {
            steps {
                checkout scm
            }
        }

        stage('Load Kubeconfig') {
            steps {
                script {
                    if (!fileExists('infra.env')) {
                        error "infra.env not found — run the k8s-infra pipeline first to provision a cluster"
                    }
                    def kubeconfigContent = readFile('infra.env').trim()
                        .split('
').find { it.startsWith('KUBECONFIG_CONTENT=') }?.split('=', 2)?.last()
                    if (!kubeconfigContent) {
                        error "KUBECONFIG_CONTENT missing from infra.env — run the k8s-infra pipeline first"
                    }
                    writeFile file: "${env.WORKSPACE}/kubeconfig-b64", text: kubeconfigContent
                    sh "base64 -d ${env.WORKSPACE}/kubeconfig-b64 > ${env.WORKSPACE}/kubeconfig && rm -f ${env.WORKSPACE}/kubeconfig-b64"
                    env.KUBECONFIG = "${env.WORKSPACE}/kubeconfig"
                }
            }
        }

        stage('Install Observability Tools') {
            when { expression { params.ACTION == 'INSTALL' } }
            steps {
                script {
                    // Metrics
                    if (params.PROMETHEUS)            { def s = load 'scripts/groovy/observability-install-prometheus.groovy';            s() }
                    if (params.ALERTMANAGER)          { def s = load 'scripts/groovy/observability-install-alertmanager.groovy';          s() }
                    if (params.THANOS)                { def s = load 'scripts/groovy/observability-install-thanos.groovy';                s() }
                    if (params.VICTORIA_METRICS)      { def s = load 'scripts/groovy/observability-install-victoria-metrics.groovy';      s() }
                    if (params.PUSHGATEWAY)           { def s = load 'scripts/groovy/observability-install-pushgateway.groovy';           s() }
                    if (params.BLACKBOX_EXPORTER)     { def s = load 'scripts/groovy/observability-install-blackbox-exporter.groovy';     s() }
                    if (params.KUBE_STATE_METRICS)    { def s = load 'scripts/groovy/observability-install-kube-state-metrics.groovy';    s() }
                    if (params.NODE_EXPORTER)         { def s = load 'scripts/groovy/observability-install-node-exporter.groovy';         s() }
                    // Logs
                    if (params.ELASTICSEARCH)         { def s = load 'scripts/groovy/observability-install-elasticsearch.groovy';         s() }
                    if (params.OPENSEARCH)            { def s = load 'scripts/groovy/observability-install-opensearch.groovy';            s() }
                    if (params.LOKI)                  { def s = load 'scripts/groovy/observability-install-loki.groovy';                  s() }
                    if (params.FLUENTD)               { def s = load 'scripts/groovy/observability-install-fluentd.groovy';               s() }
                    if (params.FLUENT_BIT)            { def s = load 'scripts/groovy/observability-install-fluent-bit.groovy';            s() }
                    if (params.LOGSTASH)              { def s = load 'scripts/groovy/observability-install-logstash.groovy';              s() }
                    // Tracing
                    if (params.JAEGER)                { def s = load 'scripts/groovy/observability-install-jaeger.groovy';                s() }
                    if (params.TEMPO)                 { def s = load 'scripts/groovy/observability-install-tempo.groovy';                 s() }
                    if (params.ZIPKIN)                { def s = load 'scripts/groovy/observability-install-zipkin.groovy';                s() }
                    // Dashboards
                    if (params.GRAFANA)               { def s = load 'scripts/groovy/observability-install-grafana.groovy';               s() }
                    if (params.KIBANA)                { def s = load 'scripts/groovy/observability-install-kibana.groovy';                s() }
                    if (params.OPENSEARCH_DASHBOARDS) { def s = load 'scripts/groovy/observability-install-opensearch-dashboards.groovy'; s() }
                    // Instrumentation
                    if (params.OTEL_COLLECTOR)        { def s = load 'scripts/groovy/observability-install-otel-collector.groovy';        s() }
                    // Error Tracking
                    if (params.SENTRY)                { def s = load 'scripts/groovy/observability-install-sentry.groovy';                s() }
                    if (params.GLITCHTIP)             { def s = load 'scripts/groovy/observability-install-glitchtip.groovy';             s() }
                    // SLO
                    if (params.PYRRA)                 { def s = load 'scripts/groovy/observability-install-pyrra.groovy';                 s() }
                    if (params.SLOTH)                 { def s = load 'scripts/groovy/observability-install-sloth.groovy';                 s() }
                    if (params.UPTIME_KUMA)           { def s = load 'scripts/groovy/observability-install-uptime-kuma.groovy';           s() }
                    // Profiling
                    if (params.PYROSCOPE)             { def s = load 'scripts/groovy/observability-install-pyroscope.groovy';             s() }
                }
            }
        }

        stage('Uninstall Observability Tools') {
            when { expression { params.ACTION == 'UNINSTALL' } }
            steps {
                script {
                    if (params.PYROSCOPE)             { sh 'helm uninstall pyroscope              -n pyroscope              --ignore-not-found || true' }
                    if (params.UPTIME_KUMA)           { sh 'helm uninstall uptime-kuma            -n uptime-kuma            --ignore-not-found || true' }
                    if (params.SLOTH)                 { sh 'helm uninstall sloth                  -n sloth                  --ignore-not-found || true' }
                    if (params.PYRRA)                 { sh 'helm uninstall pyrra                  -n pyrra                  --ignore-not-found || true' }
                    if (params.GLITCHTIP)             { sh 'helm uninstall glitchtip              -n glitchtip              --ignore-not-found || true' }
                    if (params.SENTRY)                { sh 'helm uninstall sentry                 -n sentry                 --ignore-not-found || true' }
                    if (params.OTEL_COLLECTOR)        { sh 'helm uninstall otel-collector         -n otel-collector         --ignore-not-found || true' }
                    if (params.OPENSEARCH_DASHBOARDS) { sh 'helm uninstall opensearch-dashboards  -n opensearch-dashboards  --ignore-not-found || true' }
                    if (params.KIBANA)                { sh 'helm uninstall kibana                 -n kibana                 --ignore-not-found || true' }
                    if (params.GRAFANA)               { sh 'helm uninstall grafana                -n grafana                --ignore-not-found || true' }
                    if (params.ZIPKIN)                { sh 'helm uninstall zipkin                 -n zipkin                 --ignore-not-found || true' }
                    if (params.TEMPO)                 { sh 'helm uninstall tempo                  -n tempo                  --ignore-not-found || true' }
                    if (params.JAEGER)                { sh 'helm uninstall jaeger                 -n jaeger                 --ignore-not-found || true' }
                    if (params.LOGSTASH)              { sh 'helm uninstall logstash               -n logstash               --ignore-not-found || true' }
                    if (params.FLUENT_BIT)            { sh 'helm uninstall fluent-bit             -n fluent-bit             --ignore-not-found || true' }
                    if (params.FLUENTD)               { sh 'helm uninstall fluentd                -n fluentd                --ignore-not-found || true' }
                    if (params.OPENSEARCH)            { sh 'helm uninstall opensearch             -n opensearch             --ignore-not-found || true' }
                    if (params.ELASTICSEARCH)         { sh 'helm uninstall elasticsearch          -n elasticsearch          --ignore-not-found || true' }
                    if (params.LOKI)                  { sh 'helm uninstall loki                   -n loki                   --ignore-not-found || true' }
                    if (params.NODE_EXPORTER)         { sh 'helm uninstall node-exporter          -n node-exporter          --ignore-not-found || true' }
                    if (params.KUBE_STATE_METRICS)    { sh 'helm uninstall kube-state-metrics     -n kube-state-metrics     --ignore-not-found || true' }
                    if (params.BLACKBOX_EXPORTER)     { sh 'helm uninstall blackbox-exporter      -n blackbox-exporter      --ignore-not-found || true' }
                    if (params.PUSHGATEWAY)           { sh 'helm uninstall pushgateway            -n pushgateway            --ignore-not-found || true' }
                    if (params.VICTORIA_METRICS)      { sh 'helm uninstall victoria-metrics       -n victoria-metrics       --ignore-not-found || true' }
                    if (params.THANOS)                { sh 'helm uninstall thanos                 -n thanos                 --ignore-not-found || true' }
                    if (params.ALERTMANAGER)          { sh 'helm uninstall alertmanager           -n alertmanager           --ignore-not-found || true' }
                    if (params.PROMETHEUS)            { sh 'helm uninstall prometheus             -n prometheus             --ignore-not-found || true' }
                }
            }
        }

        stage('Configure Observability Tools') {
            when { expression { params.ACTION == 'CONFIGURE' } }
            steps {
                script {
                    // Metrics
                    if (params.PROMETHEUS)            { def s = load 'scripts/groovy/observability-configure-prometheus.groovy';            s() }
                    if (params.ALERTMANAGER)          { def s = load 'scripts/groovy/observability-configure-alertmanager.groovy';          s() }
                    if (params.THANOS)                { def s = load 'scripts/groovy/observability-configure-thanos.groovy';                s() }
                    if (params.VICTORIA_METRICS)      { def s = load 'scripts/groovy/observability-configure-victoria-metrics.groovy';      s() }
                    // Logs
                    if (params.ELASTICSEARCH)         { def s = load 'scripts/groovy/observability-configure-elasticsearch.groovy';         s() }
                    if (params.OPENSEARCH)            { def s = load 'scripts/groovy/observability-configure-opensearch.groovy';            s() }
                    if (params.LOKI)                  { def s = load 'scripts/groovy/observability-configure-loki.groovy';                  s() }
                    if (params.FLUENTD)               { def s = load 'scripts/groovy/observability-configure-fluentd.groovy';               s() }
                    if (params.FLUENT_BIT)            { def s = load 'scripts/groovy/observability-configure-fluent-bit.groovy';            s() }
                    if (params.LOGSTASH)              { def s = load 'scripts/groovy/observability-configure-logstash.groovy';              s() }
                    // Tracing
                    if (params.JAEGER)                { def s = load 'scripts/groovy/observability-configure-jaeger.groovy';                s() }
                    if (params.TEMPO)                 { def s = load 'scripts/groovy/observability-configure-tempo.groovy';                 s() }
                    if (params.ZIPKIN)                { def s = load 'scripts/groovy/observability-configure-zipkin.groovy';                s() }
                    // Dashboards
                    if (params.GRAFANA)               { def s = load 'scripts/groovy/observability-configure-grafana.groovy';               s() }
                    if (params.KIBANA)                { def s = load 'scripts/groovy/observability-configure-kibana.groovy';                s() }
                    if (params.OPENSEARCH_DASHBOARDS) { def s = load 'scripts/groovy/observability-configure-opensearch-dashboards.groovy'; s() }
                    // Instrumentation
                    if (params.OTEL_COLLECTOR)        { def s = load 'scripts/groovy/observability-configure-otel-collector.groovy';        s() }
                    // Error Tracking
                    if (params.SENTRY)                { def s = load 'scripts/groovy/observability-configure-sentry.groovy';                s() }
                    if (params.GLITCHTIP)             { def s = load 'scripts/groovy/observability-configure-glitchtip.groovy';             s() }
                    // SLO
                    if (params.PYRRA)                 { def s = load 'scripts/groovy/observability-configure-pyrra.groovy';                 s() }
                    if (params.SLOTH)                 { def s = load 'scripts/groovy/observability-configure-sloth.groovy';                 s() }
                    if (params.UPTIME_KUMA)           { def s = load 'scripts/groovy/observability-configure-uptime-kuma.groovy';           s() }
                    // Profiling
                    if (params.PYROSCOPE)             { def s = load 'scripts/groovy/observability-configure-pyroscope.groovy';             s() }
                }
            }
        }
    }

    post {
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
