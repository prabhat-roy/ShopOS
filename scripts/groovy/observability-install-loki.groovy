def call() {
    sh """
        helm upgrade --install loki observability/loki/charts             --namespace loki             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^LOKI_/d' infra.env || true"
    sh "sed -i '/^LOKI_URL=/d' infra.env 2>/dev/null || true; echo 'LOKI_URL=http://loki-loki.loki.svc.cluster.local:3100' >> infra.env" 
    echo 'loki installed'
}
return this
