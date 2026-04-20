def call() {
    sh """
        helm upgrade --install harbor registry/charts/harbor \
            --namespace harbor \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://harbor-harbor.harbor.svc.cluster.local:8080'
    sh "sed -i '/^HARBOR_/d' infra.env || true"
    sh "sed -i '/^HARBOR_URL=/d' infra.env 2>/dev/null || true; echo 'HARBOR_URL=http://harbor-harbor.harbor.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^HARBOR_USER=/d' infra.env 2>/dev/null || true; echo 'HARBOR_USER=admin' >> infra.env" 
    sh "sed -i '/^HARBOR_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'HARBOR_PASSWORD=Harbor12345' >> infra.env" 

    echo 'harbor installed — ${url}'
}

return this
