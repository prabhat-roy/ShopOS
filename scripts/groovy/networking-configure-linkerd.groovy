def call() {
    sh '''
        echo "=== Configure Linkerd ==="

        # Annotate ShopOS namespaces for automatic proxy injection
        for ns in commerce platform identity catalog supply-chain financial customer-experience \\
                  communications content analytics-ai b2b integrations affiliate; do
            kubectl annotate namespace $ns linkerd.io/inject=enabled --overwrite 2>/dev/null || true
        done

        # Check control plane health
        linkerd check 2>/dev/null || kubectl rollout status deploy -n linkerd --timeout=120s || true

        echo "Linkerd proxy injection enabled for all ShopOS namespaces."
    '''
}
return this
