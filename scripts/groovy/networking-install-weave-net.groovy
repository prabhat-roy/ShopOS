def call() {
    sh '''
        helm upgrade --install weave-net networking/weave-net/charts \
            --namespace weave \
            --create-namespace \
            --set env.IPALLOC_RANGE=10.32.0.0/12 \
            --set env.WEAVE_MTU=1376 \
            --set env.NO_MASQ_LOCAL=1 \
            --set env.CHECKPOINT_DISABLE=1 \
            --set metrics.service.port=6782 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/weave-net -n weave --timeout=5m"
    sh "sed -i '/^WEAVE_/d' infra.env || true"
    sh "echo 'WEAVE_METRICS_URL=http://weave-net.weave.svc.cluster.local:6782/metrics' >> infra.env"
    echo 'Weave Net CNI installed with mesh networking and 10.32.0.0/12 IPALLOC range'
}
return this
