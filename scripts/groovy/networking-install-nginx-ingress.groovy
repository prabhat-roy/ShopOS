def call() {
    sh '''
        helm upgrade --install nginx-ingress networking/nginx-ingress/charts \
            --namespace ingress-nginx \
            --create-namespace \
            --set controller.ingressClassResource.name=nginx \
            --set controller.ingressClassResource.default=false \
            --set controller.replicaCount=2 \
            --set controller.metrics.enabled=true \
            --set controller.metrics.serviceMonitor.enabled=false \
            --set controller.podAnnotations."prometheus.io/scrape"=true \
            --set controller.podAnnotations."prometheus.io/port"=10254 \
            --set controller.config.use-real-ip=true \
            --set controller.config.forwarded-for-header="X-Forwarded-For" \
            --set controller.config.proxy-body-size=64m \
            --set controller.config.ssl-redirect=false \
            --set controller.config.use-gzip=true \
            --set controller.config.gzip-min-length=1000 \
            --set controller.allowSnippetAnnotations=false \
            --set controller.hostNetwork=false \
            --set controller.resources.requests.cpu=100m \
            --set controller.resources.requests.memory=128Mi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/nginx-ingress-ingress-nginx-controller -n ingress-nginx --timeout=5m"
    sh "sed -i '/^NGINX_INGRESS_/d' infra.env || true"
    sh "sed -i '/^NGINX_INGRESS_URL=/d' infra.env 2>/dev/null || true; echo 'NGINX_INGRESS_URL=http://nginx-ingress-ingress-nginx-controller.ingress-nginx.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^NGINX_INGRESS_CLASS=/d' infra.env 2>/dev/null || true; echo 'NGINX_INGRESS_CLASS=nginx' >> infra.env" 
    echo 'NGINX Ingress Controller installed with metrics, real-IP, and gzip compression'
}
return this
