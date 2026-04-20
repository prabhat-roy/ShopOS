def call() {
    sh '''
        helm upgrade --install anchore security/anchore/charts \
            --namespace anchore \
            --create-namespace \
            --set anchoreConfig.default_admin_password=foobar \
            --set anchoreConfig.default_admin_email=admin@shopos.local \
            --set anchoreConfig.service_dir=/anchore_service \
            --set anchoreConfig.log_level=INFO \
            --set anchoreConfig.keys.secret=anchore-seed-secret \
            --set anchoreApi.replicaCount=2 \
            --set anchoreApi.resources.requests.cpu=100m \
            --set anchoreApi.resources.requests.memory=256Mi \
            --set anchoreAnalyzer.replicaCount=2 \
            --set anchoreAnalyzer.resources.requests.cpu=250m \
            --set anchoreAnalyzer.resources.requests.memory=512Mi \
            --set anchorePolicyEngine.replicaCount=2 \
            --set anchorePolicyEngine.resources.requests.cpu=100m \
            --set anchorePolicyEngine.resources.requests.memory=256Mi \
            --set anchoreCatalog.replicaCount=2 \
            --set postgresql.enabled=true \
            --set postgresql.auth.password=anchore \
            --set postgresql.auth.database=anchore \
            --set anchoreConfig.metrics.enabled=true \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status deployment/anchore-api -n anchore --timeout=5m"
    sh "sed -i '/^ANCHORE_/d' infra.env || true"
    sh "echo 'ANCHORE_URL=http://anchore-api.anchore.svc.cluster.local:8228' >> infra.env"
    sh "echo 'ANCHORE_USER=admin' >> infra.env"
    sh "echo 'ANCHORE_PASSWORD=foobar' >> infra.env"
    echo 'Anchore Engine installed — HA API, analyzer, policy engine, catalog with PostgreSQL'
}
return this
