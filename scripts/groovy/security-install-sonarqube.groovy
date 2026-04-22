def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install sonarqube security/sonarqube/charts \
            --namespace sonarqube \
            --create-namespace \
            --set edition=community \
            --set replicaCount=1 \
            --set image.tag=10.6.0-community \
            --set resources.requests.cpu=500m \
            --set resources.requests.memory=2Gi \
            --set resources.limits.cpu=2000m \
            --set resources.limits.memory=4Gi \
            --set persistence.enabled=true \
            --set persistence.size=10Gi \
            --set persistence.storageClass=${sc} \
            --set postgresql.enabled=true \
            --set postgresql.auth.username=sonarqube \
            --set postgresql.auth.password=sonarqube \
            --set postgresql.auth.database=sonarqube \
            --set postgresql.primary.persistence.size=10Gi \
            --set ingress.enabled=false \
            --set service.type=ClusterIP \
            --set service.internalPort=9000 \
            --set monitoringPasscode=changeme \
            --set sonarProperties."sonar.forceAuthentication"=true \
            --set sonarProperties."sonar.core.serverBaseURL"=http://sonarqube.sonarqube.svc.cluster.local:9000 \
            --wait --timeout 15m
    """
    sh "kubectl rollout status statefulset/sonarqube-sonarqube -n sonarqube --timeout=10m"
    sh "sed -i '/^SONARQUBE_/d' infra.env || true"
    sh "sed -i '/^SONARQUBE_URL=/d' infra.env 2>/dev/null || true; echo 'SONARQUBE_URL=http://sonarqube-sonarqube.sonarqube.svc.cluster.local:9000' >> infra.env" 
    sh "sed -i '/^SONARQUBE_USER=/d' infra.env 2>/dev/null || true; echo 'SONARQUBE_USER=admin' >> infra.env" 
    sh "sed -i '/^SONARQUBE_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'SONARQUBE_PASSWORD=admin' >> infra.env" 
    echo 'SonarQube Community installed — forced auth, PostgreSQL, 10Gi persistence'
}
return this
