def call() {
    sh """
        helm upgrade --install nexus registry/charts/nexus \
            --namespace nexus \
            --create-namespace \
            --wait --timeout 10m
    """

    def url = 'http://nexus-nexus.nexus.svc.cluster.local:8081'
    sh "sed -i '/^NEXUS_/d' infra.env || true"
    sh "sed -i '/^NEXUS_URL=/d' infra.env 2>/dev/null || true; echo 'NEXUS_URL=http://nexus-nexus.nexus.svc.cluster.local:8081' >> infra.env" 
    sh "sed -i '/^NEXUS_USER=/d' infra.env 2>/dev/null || true; echo 'NEXUS_USER=admin' >> infra.env" 
    sh "sed -i '/^NEXUS_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'NEXUS_PASSWORD=admin123' >> infra.env" 

    echo 'nexus installed — ${url}'
}

return this
