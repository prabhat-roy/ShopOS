def call() {
    def provider = sh(script: "grep '^CLOUD_PROVIDER=' infra.env | cut -d= -f2 || echo aws", returnStdout: true).trim()
    def domainFilter = sh(script: "grep '^DOMAIN_FILTER=' infra.env | cut -d= -f2 || echo ''", returnStdout: true).trim()

    sh """
        helm upgrade --install external-dns networking/external-dns/charts \
            --namespace external-dns \
            --create-namespace \
            --set provider=${provider} \
            --set policy=upsert-only \
            --set registry=txt \
            --set txtOwnerId=shopos-k8s \
            --set txtPrefix=edns- \
            --set sources[0]=service \
            --set sources[1]=ingress \
            --set sources[2]=istio-gateway \
            --set sources[3]=istio-virtualservice \
            --set domainFilters[0]=${domainFilter ?: 'shopos.local'} \
            --set interval=1m \
            --set triggerLoopOnEvent=true \
            --set logLevel=info \
            --set logFormat=json \
            --set metrics.enabled=true \
            --set resources.requests.cpu=50m \
            --set resources.requests.memory=64Mi \
            --wait --timeout 5m
    """
    sh "kubectl rollout status deployment/external-dns -n external-dns --timeout=5m"
    sh "sed -i '/^EXTERNAL_DNS_/d' infra.env || true"
    sh "echo 'EXTERNAL_DNS_METRICS_URL=http://external-dns.external-dns.svc.cluster.local:7979/metrics' >> infra.env"
    echo "ExternalDNS installed with provider=${provider}, upsert-only policy, Istio sources"
}
return this
