def call() {
    sh """
        helm upgrade --install kube-state-metrics observability/kube-state-metrics/charts             --namespace kube-state-metrics             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^KUBE_STATE_METRICS_/d' infra.env || true"
    sh "echo 'KUBE_STATE_METRICS_URL=http://kube-state-metrics-kube-state-metrics.kube-state-metrics.svc.cluster.local:8080' >> infra.env"
    echo 'kube-state-metrics installed'
}
return this
