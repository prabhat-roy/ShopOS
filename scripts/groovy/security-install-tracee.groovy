def call() {
    sh '''
        helm upgrade --install tracee security/tracee/charts \
            --namespace tracee \
            --create-namespace \
            --set image.tag=v0.20.0 \
            --set webhook.enabled=false \
            --set config.output.options.parse-arguments=true \
            --set config.output.options.exec-env=false \
            --set config.output.options.stack-addresses=false \
            --set config.output.options.detect-syscall=true \
            --set config.output.options.exec-hash=dev-inode \
            --set config.output.json.files[0]=stdout \
            --set config.cache.cache-type=ring-buf \
            --set config.cache.mem-cache-size=512 \
            --set config.scope[0]=container \
            --set config.scope[1]=!host \
            --set resources.requests.cpu=200m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=1000m \
            --set resources.limits.memory=1024Mi \
            --set serviceMonitor.enabled=false \
            --set metrics.port=3366 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/tracee -n tracee --timeout=5m"
    sh "sed -i '/^TRACEE_/d' infra.env || true"
    sh "sed -i '/^TRACEE_METRICS_URL=/d' infra.env 2>/dev/null || true; echo 'TRACEE_METRICS_URL=http://tracee.tracee.svc.cluster.local:3366/metrics' >> infra.env" 
    echo 'Tracee installed — eBPF event collection with container scope, JSON output, ring-buf cache'
}
return this
