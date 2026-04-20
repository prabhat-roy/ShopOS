def call() {
    // Install Linkerd CRDs
    sh '''
        helm upgrade --install linkerd-crds networking/linkerd/charts/linkerd-crds \
            --namespace linkerd \
            --create-namespace \
            --wait --timeout 5m
    '''
    // Install Linkerd control plane
    sh '''
        helm upgrade --install linkerd-control-plane networking/linkerd/charts/linkerd-control-plane \
            --namespace linkerd \
            --set controllerReplicas=2 \
            --set identity.issuer.scheme=kubernetes.io/tls \
            --set podAnnotations."cluster-autoscaler.kubernetes.io/safe-to-evict"=true \
            --set proxy.resources.requests.cpu=10m \
            --set proxy.resources.requests.memory=20Mi \
            --set proxy.resources.limits.cpu=100m \
            --set proxy.resources.limits.memory=128Mi \
            --set controllerLogLevel=info \
            --set controlPlaneTracing.enabled=false \
            --set heartbeatSchedule="0 0 * * *" \
            --wait --timeout 10m
    '''
    // Install Linkerd Viz (metrics and dashboard)
    sh '''
        helm upgrade --install linkerd-viz networking/linkerd/charts/linkerd-viz \
            --namespace linkerd-viz \
            --create-namespace \
            --set dashboard.replicas=1 \
            --set grafana.enabled=true \
            --set prometheus.enabled=true \
            --set prometheusUrl="" \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/linkerd-controller -n linkerd --timeout=5m"
    sh "sed -i '/^LINKERD_/d' infra.env || true"
    sh "sed -i '/^LINKERD_URL=/d' infra.env 2>/dev/null || true; echo 'LINKERD_URL=http://linkerd-controller.linkerd.svc.cluster.local:8086' >> infra.env" 
    sh "sed -i '/^LINKERD_VIZ_URL=/d' infra.env 2>/dev/null || true; echo 'LINKERD_VIZ_URL=http://web.linkerd-viz.svc.cluster.local:8084' >> infra.env" 
    echo 'Linkerd service mesh installed — CRDs, control plane, and Viz dashboard'
}
return this
