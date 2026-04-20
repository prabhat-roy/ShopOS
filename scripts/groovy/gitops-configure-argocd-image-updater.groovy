def call() {
    sh '''
        echo "=== Configure ArgoCD Image Updater ==="

        kubectl rollout status deploy/argocd-image-updater -n argocd-image-updater --timeout=120s || true

        # Write Harbor registry credentials to the image-updater ConfigMap
        HARBOR_URL=$(grep '^HARBOR_URL=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        HARBOR_USER=$(grep '^HARBOR_USER=' infra.env 2>/dev/null | cut -d= -f2 || echo "")
        HARBOR_PASS=$(grep '^HARBOR_PASSWORD=' infra.env 2>/dev/null | cut -d= -f2 || echo "")

        if [ -n "$HARBOR_URL" ] && [ -n "$HARBOR_USER" ] && [ -n "$HARBOR_PASS" ]; then
            kubectl create secret docker-registry harbor-registry-creds \
                -n argocd-image-updater \
                --docker-server="${HARBOR_URL}" \
                --docker-username="${HARBOR_USER}" \
                --docker-password="${HARBOR_PASS}" \
                --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true
            echo "  Harbor credentials registered with ArgoCD Image Updater"
        fi

        echo "ArgoCD Image Updater configured."
    '''
}
return this
