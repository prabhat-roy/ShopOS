def call() {
    sh """
        helm upgrade --install sealed-secrets gitops/charts/sealed-secrets \
            --namespace sealed-secrets \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^SEALED_SECRETS_/d' infra.env || true"
    sh "echo 'SEALED_SECRETS_URL=http://sealed-secrets-sealed-secrets.sealed-secrets.svc.cluster.local:8080' >> infra.env"
    echo 'sealed-secrets installed'
}
return this
