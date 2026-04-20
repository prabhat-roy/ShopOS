def call() {
    sh '''
        helm upgrade --install authentik security/authentik/charts \
            --namespace authentik \
            --create-namespace \
            --set authentik.secret_key=changeme-at-least-50-chars-long-random-secret \
            --set authentik.log_level=info \
            --set authentik.avatars=initials \
            --set server.replicas=2 \
            --set server.resources.requests.cpu=100m \
            --set server.resources.requests.memory=512Mi \
            --set server.metrics.enabled=true \
            --set worker.replicas=2 \
            --set worker.resources.requests.cpu=100m \
            --set worker.resources.requests.memory=512Mi \
            --set postgresql.enabled=true \
            --set postgresql.auth.password=authentik \
            --set postgresql.auth.database=authentik \
            --set redis.enabled=true \
            --set redis.auth.enabled=false \
            --set service.type=ClusterIP \
            --set service.port=80 \
            --set blueprints.configMaps=[] \
            --set geoip.enabled=false \
            --wait --timeout 15m
    '''
    sh "kubectl rollout status deployment/authentik-server -n authentik --timeout=10m"
    sh "sed -i '/^AUTHENTIK_/d' infra.env || true"
    sh "sed -i '/^AUTHENTIK_URL=/d' infra.env 2>/dev/null || true; echo 'AUTHENTIK_URL=http://authentik.authentik.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^AUTHENTIK_USER=/d' infra.env 2>/dev/null || true; echo 'AUTHENTIK_USER=akadmin' >> infra.env" 
    sh "sed -i '/^AUTHENTIK_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'AUTHENTIK_PASSWORD=changeme' >> infra.env" 
    echo 'Authentik installed — HA server+worker, PostgreSQL, Redis, Prometheus metrics'
}
return this
