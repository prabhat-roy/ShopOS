def call() {
    sh """
        helm upgrade --install pypiserver registry/charts/pypiserver \
            --namespace pypiserver \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://pypiserver-pypiserver.pypiserver.svc.cluster.local:8080'
    sh "sed -i '/^PYPISERVER_/d' infra.env || true"
    sh "sed -i '/^PYPISERVER_URL=/d' infra.env 2>/dev/null || true; echo 'PYPISERVER_URL=http://pypiserver-pypiserver.pypiserver.svc.cluster.local:8080' >> infra.env" 

    echo 'pypiserver installed — ${url}'
}

return this
