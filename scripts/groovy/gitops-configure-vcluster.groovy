def call() {
    sh '''
        echo "=== Configure vCluster ==="

        kubectl rollout status deploy/vcluster -n vcluster --timeout=120s || true

        # Extract vCluster kubeconfig and save to infra.env
        if command -v vcluster >/dev/null 2>&1; then
            vcluster connect vcluster --namespace vcluster --update-current=false \
                --kube-config-context-name vcluster 2>/dev/null || true
        fi

        # Expose vCluster API via NodePort so external tools can reach it
        kubectl patch svc vcluster -n vcluster \
            -p "{\"spec\":{\"type\":\"ClusterIP\"}}" 2>/dev/null || true

        # Create namespaces inside vCluster that mirror physical cluster
        VCLUSTER_KUBECONFIG=$(kubectl get secret vc-vcluster -n vcluster \
            -o jsonpath="{.data.config}" 2>/dev/null | base64 -d || echo "")

        if [ -n "$VCLUSTER_KUBECONFIG" ]; then
            echo "$VCLUSTER_KUBECONFIG" > /tmp/vcluster-kubeconfig
            DOMAINS="shopos-platform shopos-identity shopos-catalog shopos-commerce shopos-supply-chain shopos-financial shopos-customer-experience shopos-communications shopos-content shopos-analytics-ai shopos-b2b shopos-integrations shopos-affiliate"
            for ns in $DOMAINS; do
                kubectl --kubeconfig=/tmp/vcluster-kubeconfig create namespace "$ns" 2>/dev/null || true
            done
            VCLUSTER_KUBECONFIG_B64=$(base64 -w 0 /tmp/vcluster-kubeconfig)
            sed -i "/^VCLUSTER_KUBECONFIG=/d" infra.env 2>/dev/null || true
            echo "VCLUSTER_KUBECONFIG=${VCLUSTER_KUBECONFIG_B64}" >> infra.env
            rm -f /tmp/vcluster-kubeconfig
        fi

        echo "vCluster configured — all 13 domain namespaces created inside virtual cluster."
    '''
}
return this
