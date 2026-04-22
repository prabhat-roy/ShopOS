def call() {
    def prometheusUrl = sh(
        script: "grep '^PROMETHEUS_URL=' infra.env 2>/dev/null | cut -d= -f2- || echo 'http://prometheus-prometheus.prometheus.svc.cluster.local:9090'",
        returnStdout: true
    ).trim()
    sh """
        helm upgrade --install opencost observability/opencost/charts \
            --namespace monitoring \
            --create-namespace \
            --set fullnameOverride=opencost \
            --set env.PROMETHEUS_SERVER_ENDPOINT=${prometheusUrl} \
            --set env.CLUSTER_ID=shopos-prod \
            --wait --timeout 5m
    """
    sh "sed -i '/^OPENCOST_/d' infra.env || true"
    sh "echo 'OPENCOST_URL=http://opencost.monitoring.svc.cluster.local:9003' >> infra.env"
    echo 'OpenCost installed'
}
return this
