def call() {
    sh """
        helm upgrade --install argo-events gitops/charts/argo-events \
            --namespace argo-events \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^ARGO_EVENTS_/d' infra.env || true"
    sh "echo 'ARGO_EVENTS_URL=http://argo-events-argo-events.argo-events.svc.cluster.local:7777' >> infra.env"
    echo 'argo-events installed'
}
return this
