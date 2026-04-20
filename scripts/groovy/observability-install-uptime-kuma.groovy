def call() {
    sh """
        helm upgrade --install uptime-kuma observability/uptime-kuma/charts             --namespace uptime-kuma             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^UPTIME_KUMA_/d' infra.env || true"
    sh "echo 'UPTIME_KUMA_URL=http://uptime-kuma-uptime-kuma.uptime-kuma.svc.cluster.local:3001' >> infra.env"
    echo 'uptime-kuma installed'
}
return this
