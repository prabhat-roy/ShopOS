def call() {
    sh '''
        echo "=== Configure Flux CD ==="

        # Verify Flux controllers are running
        kubectl rollout status deploy/source-controller     -n flux-system --timeout=120s || true
        kubectl rollout status deploy/kustomize-controller  -n flux-system --timeout=120s || true
        kubectl rollout status deploy/helm-controller       -n flux-system --timeout=120s || true

        # Bootstrap GitRepository pointing to Gitea if available
        GITEA_URL=$(grep '^GITEA_URL=' infra.env 2>/dev/null | cut -d= -f2)
        GITEA_TOKEN=$(grep '^GITEA_TOKEN=' infra.env 2>/dev/null | cut -d= -f2)

        if [ -n "$GITEA_URL" ] && [ -n "$GITEA_TOKEN" ]; then
            kubectl apply -f - <<EOF
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: enterprise-platform
  namespace: flux-system
spec:
  interval: 1m
  url: ${GITEA_URL}/shopos/enterprise-platform.git
  secretRef:
    name: gitea-credentials
  ref:
    branch: main
EOF

            # Create secret for Gitea access
            kubectl create secret generic gitea-credentials \
                -n flux-system \
                --from-literal=username=gitea \
                --from-literal=password="${GITEA_TOKEN}" \
                --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

            # Create Kustomization for the gitops/flux directory
            kubectl apply -f - <<EOF
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: shopos-platform
  namespace: flux-system
spec:
  interval: 5m
  path: ./gitops/flux
  prune: true
  sourceRef:
    kind: GitRepository
    name: enterprise-platform
EOF
            echo "  Flux GitRepository and Kustomization created."
        else
            echo "  GITEA_URL/GITEA_TOKEN not set — skipping GitRepository bootstrap"
        fi

        echo "Flux CD configuration complete."
    '''
}
return this
