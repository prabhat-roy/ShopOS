def call() {
    sh """
        helm upgrade --install fluentd observability/fluentd/charts             --namespace fluentd             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^FLUENTD_/d' infra.env || true"
    sh "sed -i '/^FLUENTD_URL=/d' infra.env 2>/dev/null || true; echo 'FLUENTD_URL=http://fluentd-fluentd.fluentd.svc.cluster.local:24224' >> infra.env" 
    echo 'fluentd installed'
}
return this
