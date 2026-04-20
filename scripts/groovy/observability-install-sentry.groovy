def call() {
    sh """
        helm upgrade --install sentry observability/sentry/charts             --namespace sentry             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^SENTRY_/d' infra.env || true"
    sh "echo 'SENTRY_URL=http://sentry-sentry.sentry.svc.cluster.local:9000' >> infra.env"
    echo 'sentry installed'
}
return this
