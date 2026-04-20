def call() {
    sh """
        LOKI_URL=\$(grep '^LOKI_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Loki at \${LOKI_URL}..."
        until curl -sf "\${LOKI_URL}/ready" > /dev/null 2>&1; do sleep 10; done

        # Load ShopOS Loki config from existing config file
        kubectl create configmap loki-config \
            --from-file=observability/loki/ \
            --namespace loki --dry-run=client -o yaml | kubectl apply -f - || true

        kubectl rollout restart deployment/loki-loki -n loki || true
        echo "Loki ready — log retention and storage configured"
    """
    echo 'loki configured'
}
return this
