def call() {
    sh '''
        echo "=== Configure External Secrets Operator ==="

        kubectl rollout status deploy/external-secrets -n external-secrets --timeout=120s || true

        # Create ClusterSecretStore pointing to Vault if available
        VAULT_URL=$(grep '^VAULT_URL=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        VAULT_TOKEN=$(grep '^VAULT_ROOT_TOKEN=' infra.env 2>/dev/null | cut -d= -f2 || echo "")

        if [ -n "$VAULT_URL" ] && [ -n "$VAULT_TOKEN" ]; then
            # Create Vault token secret
            kubectl create secret generic vault-token \
                -n external-secrets \
                --from-literal=token="${VAULT_TOKEN}" \
                --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

            kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "${VAULT_URL}"
      path: "secret"
      version: "v2"
      auth:
        tokenSecretRef:
          name: vault-token
          namespace: external-secrets
          key: token
EOF
            echo "  ClusterSecretStore vault-backend created pointing to ${VAULT_URL}"
        else
            echo "  VAULT_URL/VAULT_ROOT_TOKEN not set — skipping ClusterSecretStore creation"
        fi

        echo "External Secrets Operator configured."
    '''
}
return this
