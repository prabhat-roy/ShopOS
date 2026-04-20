def call() {
    sh '''
        helm upgrade --install kong networking/kong/charts \
            --namespace kong \
            --create-namespace \
            --set ingressController.enabled=true \
            --set ingressController.ingressClass=kong \
            --set ingressController.installCRDs=false \
            --set deployment.kong.enabled=true \
            --set env.database=off \
            --set env.router_flavor=expressions \
            --set env.nginx_worker_processes=auto \
            --set env.proxy_access_log=/dev/stdout \
            --set env.proxy_error_log=/dev/stderr \
            --set proxy.enabled=true \
            --set proxy.type=LoadBalancer \
            --set proxy.http.enabled=true \
            --set proxy.http.containerPort=8000 \
            --set proxy.tls.enabled=true \
            --set proxy.tls.containerPort=8443 \
            --set admin.enabled=true \
            --set admin.http.enabled=true \
            --set admin.http.containerPort=8001 \
            --set admin.tls.enabled=false \
            --set status.enabled=true \
            --set status.http.enabled=true \
            --set status.http.containerPort=8100 \
            --set plugins.configMaps=[] \
            --set replicaCount=2 \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=256Mi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/kong-kong -n kong --timeout=5m"
    sh "sed -i '/^KONG_/d' infra.env || true"
    sh "sed -i '/^KONG_PROXY_URL=/d' infra.env 2>/dev/null || true; echo 'KONG_PROXY_URL=http://kong-kong-proxy.kong.svc.cluster.local:8000' >> infra.env" 
    sh "sed -i '/^KONG_ADMIN_URL=/d' infra.env 2>/dev/null || true; echo 'KONG_ADMIN_URL=http://kong-kong-admin.kong.svc.cluster.local:8001' >> infra.env" 
    sh "sed -i '/^KONG_INGRESS_CLASS=/d' infra.env 2>/dev/null || true; echo 'KONG_INGRESS_CLASS=kong' >> infra.env" 
    echo 'Kong API Gateway installed in DB-less mode with Ingress Controller'
}
return this
