def call() {
    sh '''
        helm upgrade --install trivy-operator security/trivy-operator/charts \
            --namespace trivy-system \
            --create-namespace \
            --set operator.scanJobTimeout=5m \
            --set operator.concurrentScanJobsLimit=10 \
            --set operator.batchDeleteLimit=10 \
            --set operator.batchDeleteDelay=10s \
            --set operator.vulnerabilityScannerEnabled=true \
            --set operator.configAuditScannerEnabled=true \
            --set operator.rbacAssessmentScannerEnabled=true \
            --set operator.infraAssessmentScannerEnabled=true \
            --set operator.clusterComplianceEnabled=false \
            --set operator.scanNodeCollectorEnabled=true \
            --set trivy.ignoreUnfixed=false \
            --set trivy.severity=HIGH,CRITICAL \
            --set trivy.slow=true \
            --set trivy.timeout=5m \
            --set trivy.dbRepository=ghcr.io/aquasecurity/trivy-db \
            --set trivy.resources.requests.cpu=100m \
            --set trivy.resources.requests.memory=100Mi \
            --set trivy.resources.limits.cpu=500m \
            --set trivy.resources.limits.memory=500Mi \
            --set serviceMonitor.enabled=false \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=128Mi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/trivy-operator -n trivy-system --timeout=5m"
    sh "sed -i '/^TRIVY_OPERATOR_/d' infra.env || true"
    sh "echo 'TRIVY_OPERATOR_URL=http://trivy-operator.trivy-system.svc.cluster.local:80' >> infra.env"
    echo 'Trivy Operator installed — continuous vulnerability, config audit, RBAC assessment scans'
}
return this
