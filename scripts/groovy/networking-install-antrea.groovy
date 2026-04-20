def call() {
    sh '''
        helm upgrade --install antrea networking/antrea/charts \
            --namespace kube-system \
            --set antrea.featureGates.AntreaProxy=true \
            --set antrea.featureGates.EndpointSlice=true \
            --set antrea.featureGates.Traceflow=true \
            --set antrea.featureGates.NodePortLocal=true \
            --set antrea.featureGates.NetworkPolicyStats=true \
            --set antrea.featureGates.FlowExporter=true \
            --set antrea.featureGates.AntreaPolicy=true \
            --set antrea.trafficEncapMode=encap \
            --set antrea.tunnelType=geneve \
            --set antrea.serviceCIDR=10.96.0.0/12 \
            --set antrea.prometheusMetrics=true \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/antrea-agent -n kube-system --timeout=5m"
    sh "sed -i '/^ANTREA_/d' infra.env || true"
    sh "sed -i '/^ANTREA_CONTROLLER_URL=/d' infra.env 2>/dev/null || true; echo 'ANTREA_CONTROLLER_URL=https://antrea-controller.kube-system.svc.cluster.local:10349' >> infra.env" 
    echo 'Antrea CNI installed with AntreaProxy, Traceflow, and NetworkPolicyStats'
}
return this
