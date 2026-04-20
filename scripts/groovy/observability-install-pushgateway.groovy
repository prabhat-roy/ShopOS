def call() {
    sh """
        helm upgrade --install pushgateway observability/pushgateway/charts             --namespace pushgateway             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^PUSHGATEWAY_/d' infra.env || true"
    sh "sed -i '/^PUSHGATEWAY_URL=/d' infra.env 2>/dev/null || true; echo 'PUSHGATEWAY_URL=http://pushgateway-pushgateway.pushgateway.svc.cluster.local:9091' >> infra.env" 
    echo 'pushgateway installed'
}
return this
