def call() {
    sh '''
        helm upgrade --install haproxy-ingress networking/haproxy-ingress/charts \
            --namespace haproxy-ingress \
            --create-namespace \
            --set controller.ingressClass=haproxy \
            --set controller.replicaCount=2 \
            --set controller.haproxy.enabled=true \
            --set controller.metrics.enabled=true \
            --set controller.metrics.port=9101 \
            --set controller.config.timeout-connect=5s \
            --set controller.config.timeout-client=60s \
            --set controller.config.timeout-server=60s \
            --set controller.config.ssl-redirect=true \
            --set controller.config.http-server-close=true \
            --set controller.config.forwardfor=enabled \
            --set controller.config.max-connections=10000 \
            --set controller.resources.requests.cpu=100m \
            --set controller.resources.requests.memory=128Mi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/haproxy-ingress -n haproxy-ingress --timeout=5m"
    sh "sed -i '/^HAPROXY_INGRESS_/d' infra.env || true"
    sh "sed -i '/^HAPROXY_INGRESS_URL=/d' infra.env 2>/dev/null || true; echo 'HAPROXY_INGRESS_URL=http://haproxy-ingress.haproxy-ingress.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^HAPROXY_INGRESS_CLASS=/d' infra.env 2>/dev/null || true; echo 'HAPROXY_INGRESS_CLASS=haproxy' >> infra.env" 
    echo 'HAProxy Ingress Controller installed with metrics, SSL redirect, and ForwardFor'
}
return this
