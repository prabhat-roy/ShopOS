def call() {
    sh '''
        helm upgrade --install polaris security/polaris/charts \
            --namespace polaris \
            --create-namespace \
            --set dashboard.enable=true \
            --set dashboard.replicas=2 \
            --set dashboard.resources.requests.cpu=100m \
            --set dashboard.resources.requests.memory=128Mi \
            --set webhook.enable=true \
            --set webhook.replicas=2 \
            --set webhook.resources.requests.cpu=100m \
            --set webhook.resources.requests.memory=128Mi \
            --set config.checks.privileged=danger \
            --set config.checks.hostPID=danger \
            --set config.checks.hostIPC=danger \
            --set config.checks.hostNetwork=warning \
            --set config.checks.runAsRootAllowed=warning \
            --set config.checks.runAsPrivileged=danger \
            --set config.checks.notReadOnlyRootFilesystem=warning \
            --set config.checks.privilegeEscalationAllowed=danger \
            --set config.checks.dangerousCapabilities=danger \
            --set config.checks.insecureCapabilities=warning \
            --set config.checks.hostProcessSet=danger \
            --set config.checks.cpuRequestsMissing=warning \
            --set config.checks.cpuLimitsMissing=warning \
            --set config.checks.memoryRequestsMissing=warning \
            --set config.checks.memoryLimitsMissing=warning \
            --set config.checks.readinessProbeMissing=warning \
            --set config.checks.livenessProbeMissing=warning \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/polaris-dashboard -n polaris --timeout=5m"
    sh "sed -i '/^POLARIS_/d' infra.env || true"
    sh "sed -i '/^POLARIS_URL=/d' infra.env 2>/dev/null || true; echo 'POLARIS_URL=http://polaris-dashboard.polaris.svc.cluster.local:8080' >> infra.env" 
    echo 'Polaris installed — dashboard and validating webhook with security check configuration'
}
return this
