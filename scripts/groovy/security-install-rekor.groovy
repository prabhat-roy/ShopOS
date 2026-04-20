def call() {
    sh '''
        helm upgrade --install rekor security/rekor/charts \
            --namespace sigstore-system \
            --create-namespace \
            --set rekor.replicaCount=2 \
            --set rekor.image.tag=v1.3.6 \
            --set rekor.config.treeID=0 \
            --set rekor.config.enable_retrieve_api=true \
            --set rekor.config.enable_stable_checkpoint=true \
            --set rekor.resources.requests.cpu=100m \
            --set rekor.resources.requests.memory=256Mi \
            --set trillian.server.replicaCount=2 \
            --set trillian.mysql.enabled=true \
            --set trillian.mysql.replicaCount=1 \
            --set redis.enabled=true \
            --set redis.auth.enabled=false \
            --set redis.master.persistence.enabled=true \
            --set redis.master.persistence.size=1Gi \
            --set mysql.storage.size=5Gi \
            --set metrics.serviceMonitor.enabled=false \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status deployment/rekor-server -n sigstore-system --timeout=5m || true"
    sh "sed -i '/^REKOR_/d' infra.env || true"
    sh "sed -i '/^REKOR_URL=/d' infra.env 2>/dev/null || true; echo 'REKOR_URL=http://rekor-server.sigstore-system.svc.cluster.local:3000' >> infra.env" 
    echo 'Rekor transparency log installed — HA server with Trillian, MySQL, Redis for supply chain'
}
return this
