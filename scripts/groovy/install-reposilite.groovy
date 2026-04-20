def call() {
    sh """
        helm upgrade --install reposilite registry/charts/reposilite \
            --namespace reposilite \
            --create-namespace \
            --wait --timeout 5m
    """

    def url = 'http://reposilite-reposilite.reposilite.svc.cluster.local:8080'
    sh "sed -i '/^REPOSILITE_/d' infra.env || true"
    sh "sed -i '/^REPOSILITE_URL=/d' infra.env 2>/dev/null || true; echo 'REPOSILITE_URL=http://reposilite-reposilite.reposilite.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^REPOSILITE_USER=/d' infra.env 2>/dev/null || true; echo 'REPOSILITE_USER=manager' >> infra.env" 
    sh "sed -i '/^REPOSILITE_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'REPOSILITE_PASSWORD=reposilite-manager' >> infra.env" 

    echo 'reposilite installed — ${url}'
}

return this
