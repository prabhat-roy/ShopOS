def call() {
    sh '''
        helm upgrade --install tetragon security/tetragon/charts \
            --namespace kube-system \
            --set tetragon.image.tag=v1.2.0 \
            --set tetragon.resources.requests.cpu=200m \
            --set tetragon.resources.requests.memory=200Mi \
            --set tetragon.resources.limits.cpu=500m \
            --set tetragon.resources.limits.memory=500Mi \
            --set tetragon.btf="" \
            --set tetragon.prometheus.enabled=true \
            --set tetragon.prometheus.port=2112 \
            --set tetragon.prometheus.metricsLabelFilter="namespace,workload,binary,syscall" \
            --set tetragon.grpc.address=localhost:54321 \
            --set tetragonOperator.enabled=true \
            --set tetragonOperator.resources.requests.cpu=50m \
            --set tetragonOperator.resources.requests.memory=64Mi \
            --set tetragonOperator.prometheus.enabled=true \
            --set tetragonOperator.prometheus.port=2113 \
            --set export.stdout.enabled=true \
            --set export.stdout.pterm=false \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/tetragon -n kube-system --timeout=5m"
    sh "sed -i '/^TETRAGON_/d' infra.env || true"
    sh "sed -i '/^TETRAGON_GRPC_URL=/d' infra.env 2>/dev/null || true; echo 'TETRAGON_GRPC_URL=localhost:54321' >> infra.env" 
    sh "sed -i '/^TETRAGON_METRICS_URL=/d' infra.env 2>/dev/null || true; echo 'TETRAGON_METRICS_URL=http://tetragon.kube-system.svc.cluster.local:2112/metrics' >> infra.env" 
    echo 'Tetragon installed — eBPF enforcement, gRPC export, Prometheus metrics at :2112'
}
return this
