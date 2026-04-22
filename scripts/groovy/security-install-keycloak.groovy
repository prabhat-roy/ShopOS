def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install keycloak security/keycloak/charts \
            --namespace keycloak \
            --create-namespace \
            --set auth.adminUser=admin \
            --set auth.adminPassword=changeme \
            --set replicaCount=2 \
            --set production=true \
            --set proxy=edge \
            --set httpRelativePath=/ \
            --set cache.enabled=true \
            --set cache.stackName=kubernetes \
            --set postgresql.enabled=true \
            --set postgresql.auth.username=keycloak \
            --set postgresql.auth.password=keycloak \
            --set postgresql.auth.database=keycloak \
            --set postgresql.primary.resources.requests.cpu=100m \
            --set postgresql.primary.resources.requests.memory=256Mi \
            --set resources.requests.cpu=250m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=1000m \
            --set resources.limits.memory=1Gi \
            --set metrics.enabled=true \
            --set metrics.serviceMonitor.enabled=false \
            --set service.type=ClusterIP \
            --set service.ports.http=8080 \
            --set service.ports.https=8443 \
            --set keycloakConfigCli.enabled=true \
            --set extraEnvVars[0].name=KC_FEATURES \
            --set extraEnvVars[0].value=token-exchange \
            --set persistence.storageClass=${sc} \
            --wait --timeout 15m
    """
    sh "kubectl rollout status statefulset/keycloak -n keycloak --timeout=10m"
    sh "sed -i '/^KEYCLOAK_/d' infra.env || true"
    sh "sed -i '/^KEYCLOAK_URL=/d' infra.env 2>/dev/null || true; echo 'KEYCLOAK_URL=http://keycloak.keycloak.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^KEYCLOAK_USER=/d' infra.env 2>/dev/null || true; echo 'KEYCLOAK_USER=admin' >> infra.env" 
    sh "sed -i '/^KEYCLOAK_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'KEYCLOAK_PASSWORD=changeme' >> infra.env" 
    echo 'Keycloak installed — HA mode with PostgreSQL, metrics, and token-exchange feature'
}
return this
