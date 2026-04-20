def call() {
    sh '''
        helm upgrade --install kubescape security/kubescape/charts \
            --namespace kubescape \
            --create-namespace \
            --set clusterName=shopos \
            --set kubescape.enabled=true \
            --set kubescape.image.tag=v3.0.17 \
            --set kubescape.resources.requests.cpu=100m \
            --set kubescape.resources.requests.memory=256Mi \
            --set operator.enabled=true \
            --set operator.resources.requests.cpu=100m \
            --set operator.resources.requests.memory=64Mi \
            --set kollector.enabled=true \
            --set kollector.resources.requests.cpu=100m \
            --set kollector.resources.requests.memory=64Mi \
            --set nodeAgent.enabled=true \
            --set nodeAgent.resources.requests.cpu=100m \
            --set nodeAgent.resources.requests.memory=200Mi \
            --set storage.enabled=true \
            --set storage.resources.requests.cpu=100m \
            --set storage.resources.requests.memory=200Mi \
            --set synchronizer.enabled=true \
            --set serviceMonitor.enabled=false \
            --set capabilities.continuousScan=enable \
            --set capabilities.relevancy=enable \
            --set capabilities.networkPolicyService=enable \
            --set capabilities.runtimeObservability=enable \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status deployment/kubescape -n kubescape --timeout=5m"
    sh "sed -i '/^KUBESCAPE_/d' infra.env || true"
    sh "echo 'KUBESCAPE_URL=http://kubescape.kubescape.svc.cluster.local:8080' >> infra.env"
    echo 'Kubescape installed — NSA/MITRE compliance operator with continuous scan and network policy generation'
}
return this
