def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install external-secrets gitops/charts/external-secrets \
            --namespace external-secrets \
            --create-namespace \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^EXTERNAL_SECRETS_/d' infra.env || true"
    sh "sed -i '/^EXTERNAL_SECRETS_URL=/d' infra.env 2>/dev/null || true; echo 'EXTERNAL_SECRETS_URL=http://external-secrets-external-secrets.external-secrets.svc.cluster.local:8080' >> infra.env" 
    echo 'external-secrets installed'
}
return this
