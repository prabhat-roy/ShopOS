def call() {
    sh '''
        helm upgrade --install zap security/zap/charts \
            --namespace zap \
            --create-namespace \
            --set image.repository=softwaresecurityproject/zap-stable \
            --set image.tag=latest \
            --set zap.cmdLine="-daemon -host 0.0.0.0 -port 8080 -config api.disablekey=false -config api.key=zapapikey -config connection.dnsTtlSuccessfulQueries=-1" \
            --set resources.requests.cpu=500m \
            --set resources.requests.memory=512Mi \
            --set resources.limits.cpu=2000m \
            --set resources.limits.memory=2Gi \
            --set service.type=ClusterIP \
            --set service.port=8080 \
            --set persistence.enabled=true \
            --set persistence.size=2Gi \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/zap -n zap --timeout=5m"
    sh "sed -i '/^ZAP_/d' infra.env || true"
    sh "sed -i '/^ZAP_URL=/d' infra.env 2>/dev/null || true; echo 'ZAP_URL=http://zap.zap.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^ZAP_API_KEY=/d' infra.env 2>/dev/null || true; echo 'ZAP_API_KEY=zapapikey' >> infra.env" 
    echo 'OWASP ZAP installed — daemon mode, API key enabled, 2Gi persistence for session data'
}
return this
