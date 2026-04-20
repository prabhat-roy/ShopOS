def call() {
    sh '''
        helm upgrade --install contour networking/contour/charts \
            --namespace projectcontour \
            --create-namespace \
            --set contour.replicas=2 \
            --set envoy.service.type=LoadBalancer \
            --set envoy.useHostPort=false \
            --set contour.resources.requests.cpu=100m \
            --set contour.resources.requests.memory=128Mi \
            --set envoy.resources.requests.cpu=100m \
            --set envoy.resources.requests.memory=128Mi \
            --set metrics.contour.service.port=8000 \
            --set metrics.envoy.service.port=8002 \
            --set contour.ingressClass.name=contour \
            --set defaultBackend.enabled=true \
            --set contour.config.accesslog-format=envoy \
            --set contour.config.accesslog-fields[0]=req.starttime \
            --set contour.config.accesslog-fields[1]=req.method \
            --set contour.config.accesslog-fields[2]=req.path \
            --set contour.config.accesslog-fields[3]=response.code \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/contour -n projectcontour --timeout=5m"
    sh "sed -i '/^CONTOUR_/d' infra.env || true"
    sh "echo 'CONTOUR_URL=http://envoy.projectcontour.svc.cluster.local:80' >> infra.env"
    sh "echo 'CONTOUR_INGRESS_CLASS=contour' >> infra.env"
    echo 'Contour (Envoy-based) ingress installed with HTTPProxy CRD and access logging'
}
return this
