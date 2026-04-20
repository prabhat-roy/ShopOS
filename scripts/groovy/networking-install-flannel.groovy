def call() {
    sh '''
        helm upgrade --install flannel networking/flannel/charts \
            --namespace kube-flannel \
            --create-namespace \
            --set podCidr=10.244.0.0/16 \
            --set flannel.backend=vxlan \
            --set flannel.vni=1 \
            --set flannel.port=8472 \
            --set flannel.directRouting=false \
            --set serviceMonitor.enabled=false \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/kube-flannel-ds -n kube-flannel --timeout=5m"
    sh "sed -i '/^FLANNEL_/d' infra.env || true"
    sh "sed -i '/^FLANNEL_POD_CIDR=/d' infra.env 2>/dev/null || true; echo 'FLANNEL_POD_CIDR=10.244.0.0/16' >> infra.env" 
    echo 'Flannel CNI installed with VXLAN backend and 10.244.0.0/16 pod CIDR'
}
return this
