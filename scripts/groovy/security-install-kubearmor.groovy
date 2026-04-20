def call() {
    sh '''
        helm upgrade --install kubearmor security/kubearmor/charts \
            --namespace kubearmor \
            --create-namespace \
            --set kubearmor.image.tag=v1.4.0 \
            --set kubearmor.defaultFilePosture=block \
            --set kubearmor.defaultNetworkPosture=audit \
            --set kubearmor.defaultCapabilitiesPosture=audit \
            --set kubearmor.hostPID=true \
            --set kubearmor.args.grpc=32767 \
            --set kubearmor.resources.requests.cpu=100m \
            --set kubearmor.resources.requests.memory=200Mi \
            --set kubearmor.resources.limits.cpu=500m \
            --set kubearmor.resources.limits.memory=512Mi \
            --set kubearmorOperator.enabled=true \
            --set kubearmorController.enabled=true \
            --set kubearmorController.replicas=2 \
            --set kubearmorRelay.enabled=true \
            --set kubearmorRelay.image.tag=v1.4.0 \
            --set kubearmor.annotations."prometheus.io/scrape"=true \
            --set kubearmor.annotations."prometheus.io/port"=8080 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status daemonset/kubearmor -n kubearmor --timeout=5m"
    sh "sed -i '/^KUBEARMOR_/d' infra.env || true"
    sh "echo 'KUBEARMOR_GRPC_URL=kubearmor.kubearmor.svc.cluster.local:32767' >> infra.env"
    sh "echo 'KUBEARMOR_RELAY_URL=kubearmor-relay.kubearmor.svc.cluster.local:32767' >> infra.env"
    echo 'KubeArmor installed — LSM+eBPF enforcement, relay, operator, default block file posture'
}
return this
