def call() {
    sh '''
        echo "=== Configure Consul ==="

        # Enable Connect sidecar injection for ShopOS namespaces
        for ns in commerce platform identity catalog supply-chain financial customer-experience \\
                  communications content analytics-ai b2b integrations affiliate; do
            kubectl label namespace $ns consul.hashicorp.com/connect-inject=true --overwrite 2>/dev/null || true
        done

        # Default allow intention — services in same namespace can talk to each other
        kubectl exec -n consul deploy/consul-server -- consul intention create -allow "*" "*" 2>/dev/null || true

        # Write Consul HTTP API URL to infra.env
        CONSUL_IP=$(kubectl get svc consul-ui -n consul \
            -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "consul-server.consul.svc.cluster.local")
        sed -i '/^CONSUL_URL=/d' infra.env
        echo "CONSUL_URL=http://${CONSUL_IP}:8500" >> infra.env
        echo "  CONSUL_URL written to infra.env"

        echo "Consul Connect injection and default intentions configured."
    '''
}
return this
