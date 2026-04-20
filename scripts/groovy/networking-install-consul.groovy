def call() {
    sh '''
        helm upgrade --install consul networking/consul/charts \
            --namespace consul \
            --create-namespace \
            --set global.name=consul \
            --set global.datacenter=dc1 \
            --set global.image="hashicorp/consul:1.19.0" \
            --set global.enableConsulNamespaces=false \
            --set global.acls.manageSystemACLs=true \
            --set global.tls.enabled=true \
            --set global.tls.enableAutoEncrypt=true \
            --set global.metrics.enabled=true \
            --set global.metrics.enableAgentMetrics=true \
            --set global.metrics.agentMetricsRetentionTime=1m \
            --set server.enabled=true \
            --set server.replicas=3 \
            --set server.storage=10Gi \
            --set server.resources.requests.cpu=100m \
            --set server.resources.requests.memory=100Mi \
            --set server.connect=true \
            --set client.enabled=true \
            --set client.grpc=true \
            --set connectInject.enabled=true \
            --set connectInject.transparentProxy.defaultEnabled=false \
            --set ui.enabled=true \
            --set ui.service.type=ClusterIP \
            --set dns.enabled=true \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status statefulset/consul-server -n consul --timeout=5m"
    sh "sed -i '/^CONSUL_/d' infra.env || true"
    sh "echo 'CONSUL_URL=http://consul-server.consul.svc.cluster.local:8500' >> infra.env"
    sh "echo 'CONSUL_DNS_URL=consul-dns.consul.svc.cluster.local:53' >> infra.env"
    echo 'Consul installed with ACLs, TLS, Connect service mesh, metrics, and DNS'
}
return this
