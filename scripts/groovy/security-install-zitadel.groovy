def call() {
    sh '''
        helm upgrade --install zitadel security/zitadel/charts \
            --namespace zitadel \
            --create-namespace \
            --set zitadel.masterkey=MasterkeyNeedsToHave32Characters \
            --set zitadel.configmapConfig.ExternalSecure=false \
            --set zitadel.configmapConfig.ExternalDomain=zitadel.shopos.local \
            --set zitadel.configmapConfig.ExternalPort=80 \
            --set zitadel.configmapConfig.TLS.Enabled=false \
            --set zitadel.configmapConfig.FirstInstance.Org.Human.UserName=admin \
            --set replicaCount=2 \
            --set resources.requests.cpu=250m \
            --set resources.requests.memory=512Mi \
            --set pdb.enabled=true \
            --set pdb.minAvailable=1 \
            --set metrics.enabled=true \
            --set serviceMonitor.enabled=false \
            --set postgres.enabled=true \
            --set postgres.auth.database=zitadel \
            --set postgres.auth.username=zitadel \
            --set postgres.auth.password=zitadel \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status deployment/zitadel -n zitadel --timeout=5m"
    sh "sed -i '/^ZITADEL_/d' infra.env || true"
    sh "echo 'ZITADEL_URL=http://zitadel.zitadel.svc.cluster.local:8080' >> infra.env"
    sh "echo 'ZITADEL_USER=admin' >> infra.env"
    sh "echo 'ZITADEL_PASSWORD=RootPassword1!' >> infra.env"
    echo 'ZITADEL installed — 2 replicas, PostgreSQL, PDB, Prometheus metrics'
}
return this
