def call() {
    sh """
        helm upgrade --install alertmanager observability/alertmanager/charts             --namespace alertmanager             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^ALERTMANAGER_/d' infra.env || true"
    sh "sed -i '/^ALERTMANAGER_URL=/d' infra.env 2>/dev/null || true; echo 'ALERTMANAGER_URL=http://alertmanager-alertmanager.alertmanager.svc.cluster.local:9093' >> infra.env" 
    echo 'alertmanager installed'
}
return this
