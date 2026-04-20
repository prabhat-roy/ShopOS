def call() {
    sh """
        helm upgrade --install external-secrets gitops/charts/external-secrets \
            --namespace external-secrets \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^EXTERNAL_SECRETS_/d' infra.env || true"
    sh "echo 'EXTERNAL_SECRETS_URL=http://external-secrets-external-secrets.external-secrets.svc.cluster.local:8080' >> infra.env"
    echo 'external-secrets installed'
}
return this
