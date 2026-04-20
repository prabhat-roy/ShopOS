def call() {
    sh '''
        echo "=== Configure Argo Workflows ==="

        kubectl rollout status deploy/argo-server -n argo-workflows --timeout=120s || true

        # Create default service account with workflow permissions
        kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argo-workflow-sa
  namespace: argo-workflows
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: argo-workflow-sa-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: argo-role
subjects:
- kind: ServiceAccount
  name: argo-workflow-sa
  namespace: argo-workflows
EOF

        # Write Argo Workflows URL to infra.env
        ARGO_IP=$(kubectl get svc argo-server -n argo-workflows \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "argo-server.argo-workflows.svc.cluster.local")
        sed -i '/^ARGO_WORKFLOWS_URL=/d' infra.env
        echo "ARGO_WORKFLOWS_URL=https://${ARGO_IP}:2746" >> infra.env

        echo "Argo Workflows service account and URL configured."
    '''
}
return this
