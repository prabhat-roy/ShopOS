def call() {
    sh """
        helm upgrade --install nexus registry/charts/nexus \
            --namespace nexus \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://nexus-nexus.nexus.svc.cluster.local:8081'
    sh "sed -i '/^NEXUS_/d' infra.env || true"
    sh "echo 'NEXUS_URL=http://nexus-nexus.nexus.svc.cluster.local:8081' >> infra.env"
    sh "echo 'NEXUS_USER=admin' >> infra.env"
    sh "echo 'NEXUS_PASSWORD=admin123' >> infra.env"

    echo 'nexus installed — ${url}'
}

return this
