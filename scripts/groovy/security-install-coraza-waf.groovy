def call() {
    sh '''
        helm upgrade --install coraza-waf security/coraza-waf/charts \
            --namespace coraza-waf \
            --create-namespace \
            --set replicaCount=2 \
            --set image.tag=latest \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=128Mi \
            --set config.secruleengine=DetectionOnly \
            --set config.corazaRules.enabled=true \
            --set config.owaspCrs.enabled=true \
            --set config.owaspCrs.version=4.x \
            --set config.owaspCrs.anomalyInboundScore=5 \
            --set config.owaspCrs.anomalyOutboundScore=4 \
            --set config.owaspCrs.paranoia=1 \
            --set config.owaspCrs.blocking=false \
            --set config.logging.accessLog=true \
            --set config.logging.errorLog=true \
            --set config.logging.auditLog.enabled=true \
            --set config.logging.auditLog.parts="ABIJDEFHZ" \
            --set metrics.enabled=true \
            --set metrics.port=9090 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/coraza-waf -n coraza-waf --timeout=5m"
    sh "sed -i '/^CORAZA_WAF_/d' infra.env || true"
    sh "echo 'CORAZA_WAF_URL=http://coraza-waf.coraza-waf.svc.cluster.local:8080' >> infra.env"
    sh "echo 'CORAZA_WAF_METRICS_URL=http://coraza-waf.coraza-waf.svc.cluster.local:9090/metrics' >> infra.env"
    echo 'Coraza WAF installed — OWASP CRS v4 in DetectionOnly mode, paranoia level 1, audit logging'
}
return this
