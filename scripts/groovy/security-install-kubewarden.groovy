def call() {
    // Install CRDs first
    sh '''
        helm upgrade --install kubewarden-crds security/kubewarden/charts/kubewarden-crds \
            --namespace kubewarden \
            --create-namespace \
            --wait --timeout 5m
    '''
    // Install controller
    sh '''
        helm upgrade --install kubewarden-controller security/kubewarden/charts/kubewarden-controller \
            --namespace kubewarden \
            --set replicaCount=2 \
            --set resources.limits.cpu=500m \
            --set resources.limits.memory=128Mi \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=64Mi \
            --set telemetry.enabled=true \
            --set telemetry.metrics.port=8080 \
            --set telemetry.tracing.port=4317 \
            --wait --timeout 5m
    '''
    // Install default PolicyServer
    sh '''
        helm upgrade --install kubewarden-defaults security/kubewarden/charts/kubewarden-defaults \
            --namespace kubewarden \
            --set recommendedPolicies.enabled=true \
            --set recommendedPolicies.defaultPoliciesMode=monitor \
            --set policyServer.replicas=2 \
            --set policyServer.resources.limits.cpu=500m \
            --set policyServer.resources.limits.memory=512Mi \
            --set policyServer.resources.requests.cpu=100m \
            --set policyServer.resources.requests.memory=128Mi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/kubewarden-controller -n kubewarden --timeout=5m"
    sh "sed -i '/^KUBEWARDEN_/d' infra.env || true"
    sh "echo 'KUBEWARDEN_URL=http://kubewarden-controller.kubewarden.svc.cluster.local:443' >> infra.env"
    sh "echo 'KUBEWARDEN_METRICS_URL=http://kubewarden-controller.kubewarden.svc.cluster.local:8080' >> infra.env"
    echo 'Kubewarden installed — CRDs, controller, default PolicyServer with recommended policies in monitor mode'
}
return this
