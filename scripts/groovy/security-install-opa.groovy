def call() {
    sh '''
        helm upgrade --install gatekeeper security/opa/charts \
            --namespace gatekeeper-system \
            --create-namespace \
            --set replicas=3 \
            --set auditInterval=60 \
            --set auditMatchKindOnly=false \
            --set constraintViolationsLimit=20 \
            --set auditChunkSize=500 \
            --set logLevel=INFO \
            --set logMutations=false \
            --set emitAdmissionEvents=false \
            --set emitAuditEvents=false \
            --set controllerManager.resources.requests.cpu=100m \
            --set controllerManager.resources.requests.memory=256Mi \
            --set audit.resources.requests.cpu=100m \
            --set audit.resources.requests.memory=256Mi \
            --set metricsBackend=prometheus \
            --set enableExternalData=true \
            --set validatingWebhookFailurePolicy=Ignore \
            --set mutatingWebhookFailurePolicy=Ignore \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/gatekeeper-controller-manager -n gatekeeper-system --timeout=5m"
    sh "sed -i '/^OPA_/d' infra.env || true"
    sh "echo 'OPA_GATEKEEPER_URL=https://gatekeeper-webhook-service.gatekeeper-system.svc.cluster.local:443' >> infra.env"
    sh "echo 'OPA_METRICS_URL=http://gatekeeper-controller-manager.gatekeeper-system.svc.cluster.local:8888/metrics' >> infra.env"
    echo 'OPA Gatekeeper installed — 3 replicas, 60s audit interval, Prometheus metrics, external data enabled'
}
return this
