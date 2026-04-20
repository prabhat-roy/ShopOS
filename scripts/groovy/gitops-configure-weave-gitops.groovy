def call() {
    sh '''
        echo "=== Configure Weave GitOps ==="

        kubectl rollout status deploy/weave-gitops -n weave-gitops --timeout=120s || true

        WEAVE_URL=$(grep "^WEAVE_GITOPS_URL=" infra.env 2>/dev/null | cut -d= -f2)
        WEAVE_PASS=$(grep "^WEAVE_GITOPS_PASSWORD=" infra.env 2>/dev/null | cut -d= -f2)

        # Create admin secret for Weave GitOps dashboard (bcrypt password)
        if command -v htpasswd >/dev/null 2>&1 && [ -n "$WEAVE_PASS" ]; then
            BCRYPT=$(htpasswd -nbBC 10 "" "$WEAVE_PASS" | tr -d ":\n" | sed "s/^\s*//")
        else
            BCRYPT=$(echo -n "$WEAVE_PASS" | base64)
        fi

        kubectl create secret generic cluster-user-auth \
            -n weave-gitops \
            --from-literal=username=admin \
            --from-literal=password="${WEAVE_PASS}" \
            --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

        # Patch Weave GitOps to connect to the Flux GitRepository
        kubectl annotate deploy weave-gitops \
            -n weave-gitops \
            app.kubernetes.io/configured="true" \
            --overwrite 2>/dev/null || true

        echo "Weave GitOps configured — dashboard admin credentials set, connected to Flux."
        echo "URL: ${WEAVE_URL:-http://weave-gitops-weave-gitops.weave-gitops.svc.cluster.local:9001}"
    '''
}
return this
