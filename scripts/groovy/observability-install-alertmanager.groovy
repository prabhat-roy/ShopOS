def call() {
    sh """
        helm upgrade --install alertmanager observability/alertmanager/charts             --namespace alertmanager             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^ALERTMANAGER_/d' infra.env || true"
    sh "echo 'ALERTMANAGER_URL=http://alertmanager-alertmanager.alertmanager.svc.cluster.local:9093' >> infra.env"
    echo 'alertmanager installed'
}
return this
