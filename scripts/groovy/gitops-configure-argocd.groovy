def call() {
    sh '''
        echo "=== Configure ArgoCD ==="

        # Wait for ArgoCD server
        kubectl rollout status deploy/argocd-server -n argocd --timeout=180s

        ARGOCD_SERVER=$(kubectl get svc argocd-server -n argocd \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "argocd-server.argocd.svc.cluster.local")

        # Retrieve initial admin password
        ARGOCD_PASS=$(kubectl -n argocd get secret argocd-initial-admin-secret \
            -o jsonpath="{.data.password}" 2>/dev/null | base64 -d || echo "")

        # Write to infra.env
        sed -i '/^ARGOCD_URL=/d; /^ARGOCD_ADMIN_PASS=/d' infra.env
        echo "ARGOCD_URL=https://${ARGOCD_SERVER}" >> infra.env
        [ -n "$ARGOCD_PASS" ] && echo "ARGOCD_ADMIN_PASS=${ARGOCD_PASS}" >> infra.env

        # Login with argocd CLI if available
        if command -v argocd >/dev/null 2>&1 && [ -n "$ARGOCD_PASS" ]; then
            argocd login "${ARGOCD_SERVER}" \
                --username admin \
                --password "${ARGOCD_PASS}" \
                --insecure 2>/dev/null || true

            # Connect Gitea/GitHub repo if GITEA_URL is set
            GITEA_URL=$(grep '^GITEA_URL=' infra.env 2>/dev/null | cut -d= -f2)
            GITEA_TOKEN=$(grep '^GITEA_TOKEN=' infra.env 2>/dev/null | cut -d= -f2)
            if [ -n "$GITEA_URL" ] && [ -n "$GITEA_TOKEN" ]; then
                argocd repo add "${GITEA_URL}/shopos/enterprise-platform.git" \
                    --username gitea \
                    --password "${GITEA_TOKEN}" \
                    --insecure-skip-server-verification 2>/dev/null || true
                echo "  Git repo registered with ArgoCD"
            fi

            # Create App-of-Apps pointing to gitops/argocd/
            argocd app create app-of-apps \
                --repo "${GITEA_URL}/shopos/enterprise-platform.git" \
                --path gitops/argocd \
                --dest-server https://kubernetes.default.svc \
                --dest-namespace argocd \
                --sync-policy automated \
                --auto-prune \
                --self-heal 2>/dev/null || true
        fi

        echo "ArgoCD configured — admin password written to infra.env, app-of-apps created."
    '''
}
return this
