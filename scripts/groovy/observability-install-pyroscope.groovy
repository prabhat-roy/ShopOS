def call() {
    sh """
        helm upgrade --install pyroscope observability/pyroscope/charts             --namespace pyroscope             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^PYROSCOPE_/d' infra.env || true"
    sh "echo 'PYROSCOPE_URL=http://pyroscope-pyroscope.pyroscope.svc.cluster.local:4040' >> infra.env"
    echo 'pyroscope installed'
}
return this
