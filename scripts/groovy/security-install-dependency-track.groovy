def call() {
    sh '''
        helm upgrade --install dependency-track security/dependency-track/charts \
            --namespace dependency-track \
            --create-namespace \
            --set apiServer.replicaCount=2 \
            --set apiServer.image.tag=4.11.0 \
            --set apiServer.resources.requests.cpu=500m \
            --set apiServer.resources.requests.memory=2Gi \
            --set apiServer.resources.limits.cpu=2000m \
            --set apiServer.resources.limits.memory=4Gi \
            --set apiServer.persistentVolume.enabled=true \
            --set apiServer.persistentVolume.size=2Gi \
            --set frontend.replicaCount=2 \
            --set frontend.image.tag=4.11.0 \
            --set frontend.resources.requests.cpu=100m \
            --set frontend.resources.requests.memory=64Mi \
            --set postgresql.enabled=true \
            --set postgresql.auth.username=dtrack \
            --set postgresql.auth.password=dtrack \
            --set postgresql.auth.database=dtrack \
            --set postgresql.primary.persistence.size=10Gi \
            --set serviceMonitor.enabled=false \
            --wait --timeout 15m
    '''
    sh "kubectl rollout status deployment/dependency-track-apiserver -n dependency-track --timeout=10m"
    sh "sed -i '/^DEPENDENCY_TRACK_/d' infra.env || true"
    sh "echo 'DEPENDENCY_TRACK_URL=http://dependency-track-apiserver.dependency-track.svc.cluster.local:8080' >> infra.env"
    sh "echo 'DEPENDENCY_TRACK_FRONTEND_URL=http://dependency-track-frontend.dependency-track.svc.cluster.local:8080' >> infra.env"
    sh "echo 'DEPENDENCY_TRACK_USER=admin' >> infra.env"
    sh "echo 'DEPENDENCY_TRACK_PASSWORD=admin' >> infra.env"
    echo 'OWASP Dependency-Track installed — HA API server, frontend, PostgreSQL for SBOM management'
}
return this
