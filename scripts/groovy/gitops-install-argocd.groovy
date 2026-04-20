def call() {
    sh """
        helm upgrade --install argocd gitops/charts/argocd \
            --namespace argocd \
            --create-namespace \
            --wait --timeout 10m
    """
    sh "sed -i '/^ARGOCD_/d' infra.env || true"
    sh "sed -i '/^ARGOCD_URL=/d' infra.env 2>/dev/null || true; echo 'ARGOCD_URL=http://argocd-argocd.argocd.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^ARGOCD_USER=/d' infra.env 2>/dev/null || true; echo 'ARGOCD_USER=admin' >> infra.env" 
    sh """
        ARGOCD_PWD=\$(kubectl get secret argocd-initial-admin-secret \
            -n argocd -o jsonpath='{.data.password}' | base64 -d 2>/dev/null || echo 'see-argocd-secret')
        echo "ARGOCD_PASSWORD=\${ARGOCD_PWD}" >> infra.env
    """
    echo 'argocd installed'
}
return this
