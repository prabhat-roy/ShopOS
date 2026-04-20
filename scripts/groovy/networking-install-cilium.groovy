def call() {
    sh '''
        helm upgrade --install cilium networking/cilium/charts \
            --namespace cilium \
            --create-namespace \
            --set kubeProxyReplacement=strict \
            --set k8sServiceHost=$(kubectl get endpoints kubernetes -o jsonpath='{.subsets[0].addresses[0].ip}') \
            --set k8sServicePort=6443 \
            --set hubble.relay.enabled=true \
            --set hubble.ui.enabled=true \
            --set hubble.metrics.enableOpenMetrics=true \
            --set hubble.metrics.enabled="{dns,drop,tcp,flow,port-distribution,icmp,httpV2:exemplars=true;labelsContext=source_ip\\,source_namespace\\,source_workload\\,destination_ip\\,destination_namespace\\,destination_workload\\,traffic_direction}" \
            --set operator.replicas=2 \
            --set ipam.mode=kubernetes \
            --set bandwidthManager.enabled=true \
            --set loadBalancer.algorithm=maglev \
            --set l7Proxy=true \
            --set encryption.enabled=false \
            --set prometheus.enabled=true \
            --set operator.prometheus.enabled=true \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status daemonset/cilium -n cilium --timeout=5m"
    sh "sed -i '/^CILIUM_/d' infra.env || true"
    sh "echo 'CILIUM_HUBBLE_URL=http://hubble-relay.cilium.svc.cluster.local:80' >> infra.env"
    sh "echo 'CILIUM_METRICS_URL=http://cilium.cilium.svc.cluster.local:9962/metrics' >> infra.env"
    echo 'Cilium CNI installed with Hubble observability and kube-proxy replacement'
}
return this
