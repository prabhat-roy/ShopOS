def call() {
    sh """
        helm upgrade --install verdaccio registry/charts/verdaccio \
            --namespace verdaccio \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://verdaccio-verdaccio.verdaccio.svc.cluster.local:4873'
    sh "sed -i '/^VERDACCIO_/d' infra.env || true"
    sh "sed -i '/^VERDACCIO_URL=/d' infra.env 2>/dev/null || true; echo 'VERDACCIO_URL=http://verdaccio-verdaccio.verdaccio.svc.cluster.local:4873' >> infra.env" 

    echo 'verdaccio installed — ${url}'
}

return this
