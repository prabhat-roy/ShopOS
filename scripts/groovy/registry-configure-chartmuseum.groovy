def call() {
    sh '''
        echo "=== Configure ChartMuseum ==="

        kubectl rollout status deploy/chartmuseum -n chartmuseum --timeout=120s || true

        CM_IP=$(kubectl get svc chartmuseum -n chartmuseum \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "chartmuseum.chartmuseum.svc.cluster.local")
        CM_URL="http://${CM_IP}:8080"

        # Add ChartMuseum as a Helm repo on the Jenkins node
        helm repo add shopos-charts "${CM_URL}" 2>/dev/null || true
        helm repo update 2>/dev/null || true

        sed -i '/^CHARTMUSEUM_URL=/d' infra.env
        echo "CHARTMUSEUM_URL=${CM_URL}" >> infra.env
        echo "  ChartMuseum URL written to infra.env and added as helm repo 'shopos-charts'"

        echo "ChartMuseum configured."
    '''
}
return this
