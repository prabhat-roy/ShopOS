def call() {
    sh '''
        helm upgrade --install traefik networking/traefik/charts \
            --namespace traefik \
            --create-namespace \
            --set ingressClass.enabled=true \
            --set ingressClass.isDefaultClass=true \
            --set ingressClass.name=traefik \
            --set deployment.replicas=2 \
            --set ports.web.port=8000 \
            --set ports.web.exposedPort=80 \
            --set ports.websecure.port=8443 \
            --set ports.websecure.exposedPort=443 \
            --set ports.web.redirectTo.port=websecure \
            --set dashboard.enabled=true \
            --set api.insecure=false \
            --set api.dashboard=true \
            --set metrics.prometheus.enabled=true \
            --set metrics.prometheus.addEntryPointsLabels=true \
            --set metrics.prometheus.addRoutersLabels=true \
            --set metrics.prometheus.addServicesLabels=true \
            --set logs.general.level=INFO \
            --set logs.access.enabled=true \
            --set providers.kubernetesIngress.enabled=true \
            --set providers.kubernetesCRD.enabled=true \
            --set providers.kubernetesIngress.publishedService.enabled=true \
            --set globalArguments="{--global.sendAnonymousUsage=false}" \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=128Mi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/traefik -n traefik --timeout=5m"
    sh "sed -i '/^TRAEFIK_/d' infra.env || true"
    sh "sed -i '/^TRAEFIK_URL=/d' infra.env 2>/dev/null || true; echo 'TRAEFIK_URL=http://traefik.traefik.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^TRAEFIK_DASHBOARD_URL=/d' infra.env 2>/dev/null || true; echo 'TRAEFIK_DASHBOARD_URL=http://traefik.traefik.svc.cluster.local:9000/dashboard/' >> infra.env" 
    sh "sed -i '/^TRAEFIK_INGRESS_CLASS=/d' infra.env 2>/dev/null || true; echo 'TRAEFIK_INGRESS_CLASS=traefik' >> infra.env" 
    echo 'Traefik installed as default IngressClass with dashboard and Prometheus metrics'
}
return this
