def call() {
    sh """
        helm upgrade --install terrareg registry/charts/terrareg \
            --namespace terrareg \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://terrareg-terrareg.terrareg.svc.cluster.local:5000'
    sh "sed -i '/^TERRAREG_/d' infra.env || true"
    sh "sed -i '/^TERRAREG_URL=/d' infra.env 2>/dev/null || true; echo 'TERRAREG_URL=http://terrareg-terrareg.terrareg.svc.cluster.local:5000' >> infra.env" 

    echo 'terrareg installed — ${url}'
}

return this
