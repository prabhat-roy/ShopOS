def call() {
    sh '''
        echo "=== Configure Flux CD ==="

        kubectl rollout status deploy/source-controller    -n flux-system --timeout=120s || true
        kubectl rollout status deploy/kustomize-controller -n flux-system --timeout=120s || true
        kubectl rollout status deploy/helm-controller      -n flux-system --timeout=120s || true
        kubectl rollout status deploy/notification-controller -n flux-system --timeout=120s || true

        # Primary source: GitHub (always available)
        kubectl apply -f - <<EOF
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: shopos
  namespace: flux-system
spec:
  interval: 1m
  url: https://github.com/prabhat-roy/ShopOS.git
  ref:
    branch: main
EOF

        # Kustomization for infrastructure layer (namespaces, RBAC, network policies)
        kubectl apply -f - <<EOF
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: shopos-infra
  namespace: flux-system
spec:
  interval: 5m
  path: ./gitops/flux/base
  prune: true
  sourceRef:
    kind: GitRepository
    name: shopos
EOF

        # Kustomization for production overlay (3 replicas, production resource limits)
        kubectl apply -f - <<EOF
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: shopos-production
  namespace: flux-system
spec:
  interval: 5m
  path: ./gitops/flux/clusters/production
  prune: true
  dependsOn:
    - name: shopos-infra
  sourceRef:
    kind: GitRepository
    name: shopos
EOF

        # Optional: Gitea mirror (use if self-hosted Gitea is available)
        GITEA_URL=$(grep "^GITEA_URL=" infra.env 2>/dev/null | cut -d= -f2)
        GITEA_TOKEN=$(grep "^GITEA_TOKEN=" infra.env 2>/dev/null | cut -d= -f2)
        if [ -n "$GITEA_URL" ] && [ -n "$GITEA_TOKEN" ]; then
            kubectl create secret generic gitea-credentials \
                -n flux-system \
                --from-literal=username=gitea \
                --from-literal=password="${GITEA_TOKEN}" \
                --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

            kubectl apply -f - <<EOF
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: shopos-gitea
  namespace: flux-system
spec:
  interval: 1m
  url: ${GITEA_URL}/shopos/ShopOS.git
  secretRef:
    name: gitea-credentials
  ref:
    branch: main
EOF
            echo "  Gitea mirror GitRepository created."
        fi

        echo "Flux CD configured — GitRepository + Kustomizations applied."
    '''
}
return this
