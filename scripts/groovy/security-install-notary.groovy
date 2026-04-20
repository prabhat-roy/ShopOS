def call() {
    sh '''
        helm upgrade --install notation security/notary/charts \
            --namespace notation \
            --create-namespace \
            --set server.replicaCount=2 \
            --set server.resources.requests.cpu=100m \
            --set server.resources.requests.memory=128Mi \
            --set signer.replicaCount=2 \
            --set signer.resources.requests.cpu=100m \
            --set signer.resources.requests.memory=128Mi \
            --set postgresql.enabled=true \
            --set postgresql.auth.database=notary \
            --set postgresql.auth.username=notary \
            --set postgresql.auth.password=notary \
            --set serviceMonitor.enabled=false \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/notation-server -n notation --timeout=5m || true"
    sh "sed -i '/^NOTARY_/d' infra.env || true"
    sh "echo 'NOTARY_URL=https://notation.notation.svc.cluster.local:443' >> infra.env"
    echo 'Notary/Notation installed — image signing service with PostgreSQL storage'
}
return this
