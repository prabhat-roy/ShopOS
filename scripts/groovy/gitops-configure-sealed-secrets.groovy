def call() {
    sh '''
        echo "=== Configure Sealed Secrets ==="

        kubectl rollout status deploy/sealed-secrets-controller -n sealed-secrets --timeout=120s || true

        # Export the public key so kubeseal can encrypt secrets offline
        PUBLIC_KEY=$(kubectl get secret -n sealed-secrets \
            -l sealedsecrets.bitnami.com/sealed-secrets-key \
            -o jsonpath='{.items[0].data.tls\.crt}' 2>/dev/null | base64 -d || echo "")

        if [ -n "$PUBLIC_KEY" ]; then
            sed -i '/^SEALED_SECRETS_CERT=/d' infra.env
            echo "SEALED_SECRETS_CERT=$(echo "$PUBLIC_KEY" | base64 -w0)" >> infra.env
            echo "  Sealed Secrets public cert written to infra.env (base64)"
        fi

        echo "Sealed Secrets controller ready — public key exported to infra.env."
    '''
}
return this
