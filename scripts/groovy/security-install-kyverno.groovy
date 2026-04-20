def call() {
    sh '''
        helm upgrade --install kyverno security/kyverno/charts \
            --namespace kyverno \
            --create-namespace \
            --set admissionController.replicas=3 \
            --set backgroundController.replicas=2 \
            --set cleanupController.replicas=2 \
            --set reportsController.replicas=2 \
            --set admissionController.resources.requests.cpu=100m \
            --set admissionController.resources.requests.memory=128Mi \
            --set backgroundController.resources.requests.cpu=100m \
            --set backgroundController.resources.requests.memory=128Mi \
            --set admissionController.serviceMonitor.enabled=false \
            --set config.webhooks[0].namespaceSelector.matchExpressions[0].key=kubernetes.io/metadata.name \
            --set config.webhooks[0].namespaceSelector.matchExpressions[0].operator=NotIn \
            --set "config.webhooks[0].namespaceSelector.matchExpressions[0].values={kube-system,kyverno}" \
            --set features.policyExceptions.enabled=true \
            --set features.generateValidatingAdmissionPolicy.enabled=false \
            --set cleanupJobs.admissionReports.enabled=true \
            --set cleanupJobs.clusterAdmissionReports.enabled=true \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/kyverno-admission-controller -n kyverno --timeout=5m"
    sh "sed -i '/^KYVERNO_/d' infra.env || true"
    sh "sed -i '/^KYVERNO_WEBHOOK_URL=/d' infra.env 2>/dev/null || true; echo 'KYVERNO_WEBHOOK_URL=https://kyverno-svc.kyverno.svc.cluster.local:443' >> infra.env" 
    sh "sed -i '/^KYVERNO_METRICS_URL=/d' infra.env 2>/dev/null || true; echo 'KYVERNO_METRICS_URL=http://kyverno-svc-metrics.kyverno.svc.cluster.local:8000' >> infra.env" 
    echo 'Kyverno installed — HA admission controller, background/cleanup/reports controllers'
}
return this
