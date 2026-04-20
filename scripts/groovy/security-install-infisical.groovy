def call() {
    sh '''
        helm upgrade --install infisical security/infisical/charts \
            --namespace infisical \
            --create-namespace \
            --set infisical.replicaCount=2 \
            --set infisical.image.tag=v0.93.0 \
            --set infisical.autoDatabaseSchemaMigration=true \
            --set infisical.resources.requests.cpu=100m \
            --set infisical.resources.requests.memory=256Mi \
            --set infisical.kubeSecretRef=infisical-secrets \
            --set mongodb.enabled=true \
            --set mongodb.auth.enabled=false \
            --set mongodb.persistence.enabled=true \
            --set mongodb.persistence.size=10Gi \
            --set redis.enabled=true \
            --set redis.auth.enabled=false \
            --set ingress.enabled=false \
            --set mailhog.enabled=false \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status deployment/infisical -n infisical --timeout=5m"
    sh "sed -i '/^INFISICAL_/d' infra.env || true"
    sh "echo 'INFISICAL_URL=http://infisical.infisical.svc.cluster.local:8080' >> infra.env"
    echo 'Infisical installed — 2 replicas, MongoDB persistence, Redis, auto schema migration'
}
return this
