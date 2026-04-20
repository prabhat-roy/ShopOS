def call() {
    sh '''
        helm upgrade --install calico networking/calico/charts \
            --namespace calico-system \
            --create-namespace \
            --set installation.cni.type=Calico \
            --set installation.calicoNetwork.ipPools[0].cidr=192.168.0.0/16 \
            --set installation.calicoNetwork.ipPools[0].encapsulation=VXLAN \
            --set installation.calicoNetwork.ipPools[0].natOutgoing=Enabled \
            --set installation.calicoNetwork.bgp=Disabled \
            --set installation.controlPlaneReplicas=2 \
            --set installation.typhaMetricsPort=9093 \
            --set installation.nodeMetricsPort=9091 \
            --set installation.variant=Calico \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status daemonset/calico-node -n calico-system --timeout=5m"
    sh "sed -i '/^CALICO_/d' infra.env || true"
    sh "echo 'CALICO_NODE_METRICS_URL=http://calico-node.calico-system.svc.cluster.local:9091' >> infra.env"
    sh "echo 'CALICO_TYPHA_URL=http://calico-typha.calico-system.svc.cluster.local:9093' >> infra.env"
    echo 'Calico CNI installed with VXLAN encapsulation and NetworkPolicy enforcement'
}
return this
