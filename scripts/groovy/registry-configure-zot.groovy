def call() {
    sh '''
        echo "=== Configure Zot ==="

        kubectl rollout status deploy/zot -n zot --timeout=120s || true

        ZOT_IP=$(kubectl get svc zot -n zot \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "zot.zot.svc.cluster.local")
        sed -i '/^ZOT_URL=/d' infra.env
        echo "ZOT_URL=http://${ZOT_IP}:5000" >> infra.env
        echo "  Zot registry URL written to infra.env."

        echo "Zot OCI registry ready."
    '''
}
return this
