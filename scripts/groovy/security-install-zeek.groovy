def call() {
    sh '''
        helm upgrade --install zeek security/zeek/charts \
            --namespace zeek \
            --create-namespace \
            --set zeek.config.interfaces[0]=eth0 \
            --set zeek.config.logRotationInterval=3600 \
            --set zeek.config.logRotationSize=0 \
            --set zeek.scripts.enabled=true \
            --set zeek.scripts.packages[0]=zeek-af_packet \
            --set zeek.scripts.packages[1]=json-streaming-logs \
            --set zeek.config.JsonStreaming=true \
            --set zeek.outputs.json=true \
            --set zeek.outputs.stdout=true \
            --set zeek.config.SitePolicyScripts[0]=local.zeek \
            --set resources.requests.cpu=200m \
            --set resources.requests.memory=256Mi \
            --set resources.limits.cpu=1000m \
            --set resources.limits.memory=1Gi \
            --set metrics.enabled=true \
            --set metrics.port=47761 \
            --set persistence.enabled=true \
            --set persistence.size=10Gi \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status daemonset/zeek -n zeek --timeout=5m"
    sh "sed -i '/^ZEEK_/d' infra.env || true"
    sh "sed -i '/^ZEEK_METRICS_URL=/d' infra.env 2>/dev/null || true; echo 'ZEEK_METRICS_URL=http://zeek.zeek.svc.cluster.local:47761/metrics' >> infra.env" 
    echo 'Zeek network security monitor installed — DaemonSet, JSON streaming logs, AF_PACKET'
}
return this
