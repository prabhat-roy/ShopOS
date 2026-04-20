def call() {
    sh '''
        echo "=== Configure Forgejo ==="

        kubectl rollout status deploy/forgejo -n forgejo --timeout=120s || true

        FORGEJO_IP=$(kubectl get svc forgejo -n forgejo \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "forgejo.forgejo.svc.cluster.local")
        sed -i '/^FORGEJO_URL=/d' infra.env
        echo "FORGEJO_URL=http://${FORGEJO_IP}:3000" >> infra.env

        echo "Forgejo URL written to infra.env."
    '''
}
return this
