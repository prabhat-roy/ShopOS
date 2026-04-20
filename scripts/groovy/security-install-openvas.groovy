def call() {
    sh '''
        helm upgrade --install openvas security/openvas/charts \
            --namespace openvas \
            --create-namespace \
            --set image.tag=22.4 \
            --set gvmd.resources.requests.cpu=500m \
            --set gvmd.resources.requests.memory=512Mi \
            --set gvmd.resources.limits.cpu=2000m \
            --set gvmd.resources.limits.memory=2Gi \
            --set gsad.resources.requests.cpu=100m \
            --set gsad.resources.requests.memory=128Mi \
            --set ospd.resources.requests.cpu=500m \
            --set ospd.resources.requests.memory=512Mi \
            --set persistence.enabled=true \
            --set persistence.size=20Gi \
            --set postgresql.enabled=true \
            --set postgresql.auth.database=gvmd \
            --set postgresql.auth.username=gvmd \
            --set postgresql.auth.password=gvmd \
            --set service.type=ClusterIP \
            --set service.port=9392 \
            --set adminUser.password=admin \
            --set nvt.sync.enabled=true \
            --wait --timeout 20m
    '''
    sh "kubectl rollout status deployment/openvas-gsad -n openvas --timeout=10m || true"
    sh "sed -i '/^OPENVAS_/d' infra.env || true"
    sh "echo 'OPENVAS_URL=http://openvas.openvas.svc.cluster.local:9392' >> infra.env"
    sh "echo 'OPENVAS_USER=admin' >> infra.env"
    sh "echo 'OPENVAS_PASSWORD=admin' >> infra.env"
    echo 'OpenVAS/Greenbone installed — GVMD, GSAD, OSPD scanner with NVT sync and PostgreSQL'
}
return this
