def call() {
    // Install Istio base (CRDs)
    sh '''
        helm upgrade --install istio-base networking/istio/charts/base \
            --namespace istio-system \
            --create-namespace \
            --set defaultRevision=default \
            --wait --timeout 5m
    '''
    // Install istiod (control plane)
    sh '''
        helm upgrade --install istiod networking/istio/charts/istiod \
            --namespace istio-system \
            --set pilot.traceSampling=1.0 \
            --set pilot.resources.requests.cpu=100m \
            --set pilot.resources.requests.memory=512Mi \
            --set global.proxy.resources.requests.cpu=10m \
            --set global.proxy.resources.requests.memory=40Mi \
            --set global.proxy.resources.limits.cpu=100m \
            --set global.proxy.resources.limits.memory=128Mi \
            --set global.proxy.holdApplicationUntilProxyStarts=true \
            --set global.tracer.zipkin.address=zipkin.observability.svc.cluster.local:9411 \
            --set meshConfig.enableTracing=true \
            --set meshConfig.defaultConfig.tracing.sampling=1.0 \
            --set meshConfig.accessLogFile=/dev/stdout \
            --set meshConfig.accessLogFormat="" \
            --set meshConfig.outboundTrafficPolicy.mode=ALLOW_ANY \
            --set meshConfig.defaultConfig.proxyMetadata.BOOTSTRAP_XDS_AGENT=true \
            --wait --timeout 10m
    '''
    // Install Istio Ingress Gateway
    sh '''
        helm upgrade --install istio-ingressgateway networking/istio/charts/gateway \
            --namespace istio-system \
            --set service.type=LoadBalancer \
            --set autoscaling.enabled=true \
            --set autoscaling.minReplicas=2 \
            --set autoscaling.maxReplicas=5 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/istiod -n istio-system --timeout=5m"
    sh "sed -i '/^ISTIO_/d' infra.env || true"
    sh "echo 'ISTIO_URL=http://istiod.istio-system.svc.cluster.local:15010' >> infra.env"
    sh "echo 'ISTIO_PILOT_URL=http://istiod.istio-system.svc.cluster.local:15014' >> infra.env"
    echo 'Istio service mesh installed — base CRDs, istiod control plane, ingress gateway'
}
return this
