def call() {
    sh '''
        helm upgrade --install defectdojo security/defectdojo/charts \
            --namespace defectdojo \
            --create-namespace \
            --set django.ingress.enabled=false \
            --set django.replicas=2 \
            --set django.resources.requests.cpu=250m \
            --set django.resources.requests.memory=256Mi \
            --set django.resources.limits.cpu=1000m \
            --set django.resources.limits.memory=1Gi \
            --set celery.beat.resources.requests.cpu=100m \
            --set celery.beat.resources.requests.memory=128Mi \
            --set celery.worker.replicas=2 \
            --set celery.worker.resources.requests.cpu=100m \
            --set celery.worker.resources.requests.memory=256Mi \
            --set postgresql.enabled=true \
            --set postgresql.auth.username=defectdojo \
            --set postgresql.auth.password=defectdojo \
            --set postgresql.auth.database=defectdojo \
            --set postgresql.primary.persistence.size=5Gi \
            --set redis.enabled=true \
            --set redis.auth.enabled=false \
            --set django.initialize=true \
            --set createSecret=true \
            --set admin.user=admin \
            --set admin.password=defectdojo \
            --set admin.email=admin@shopos.local \
            --set host=defectdojo.defectdojo.svc.cluster.local \
            --wait --timeout 20m
    '''
    sh "kubectl rollout status deployment/defectdojo-django -n defectdojo --timeout=15m"
    sh "sed -i '/^DEFECTDOJO_/d' infra.env || true"
    sh "sed -i '/^DEFECTDOJO_URL=/d' infra.env 2>/dev/null || true; echo 'DEFECTDOJO_URL=http://defectdojo.defectdojo.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^DEFECTDOJO_USER=/d' infra.env 2>/dev/null || true; echo 'DEFECTDOJO_USER=admin' >> infra.env" 
    sh "sed -i '/^DEFECTDOJO_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'DEFECTDOJO_PASSWORD=defectdojo' >> infra.env" 
    echo 'DefectDojo installed — HA Django+Celery workers, PostgreSQL, Redis for vulnerability management'
}
return this
