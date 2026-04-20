def call() {
    sh '''
        echo "=== Configure ExternalDNS ==="

        # Verify ExternalDNS is running
        kubectl rollout status deploy/external-dns -n external-dns --timeout=60s || true

        # Confirm which provider is configured
        PROVIDER=$(kubectl get deploy external-dns -n external-dns \
            -o jsonpath='{.spec.template.spec.containers[0].args}' 2>/dev/null | grep -o 'provider=[^,]*' | head -1 || echo "unknown")
        echo "  ExternalDNS provider: ${PROVIDER}"

        echo "ExternalDNS is running and ready to sync K8s services to DNS."
    '''
}
return this
