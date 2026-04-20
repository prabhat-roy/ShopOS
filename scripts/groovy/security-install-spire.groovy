def call() {
    sh '''
        helm upgrade --install spire security/spire/charts \
            --namespace spire-system \
            --create-namespace \
            --set global.spire.trustDomain=shopos.cluster.local \
            --set global.spire.clusterName=shopos \
            --set spire-server.replicaCount=3 \
            --set spire-server.dataStorage.enabled=true \
            --set spire-server.dataStorage.size=1Gi \
            --set spire-server.ca.subject.country=US \
            --set spire-server.ca.subject.organization=ShopOS \
            --set spire-server.ca.subject.commonName=shopos-spire-ca \
            --set spire-server.resources.requests.cpu=100m \
            --set spire-server.resources.requests.memory=128Mi \
            --set spire-agent.resources.requests.cpu=50m \
            --set spire-agent.resources.requests.memory=64Mi \
            --set spiffe-csi-driver.enabled=true \
            --set spiffe-oidc-discovery-provider.enabled=true \
            --set spiffe-oidc-discovery-provider.config.domains[0]=oidc-discovery.shopos.cluster.local \
            --set tornjak-frontend.enabled=false \
            --wait --timeout 10m
    '''
    sh "kubectl rollout status statefulset/spire-server -n spire-system --timeout=5m"
    sh "sed -i '/^SPIRE_/d' infra.env || true"
    sh "echo 'SPIRE_SERVER_URL=spire-server.spire-system.svc.cluster.local:8081' >> infra.env"
    sh "echo 'SPIRE_TRUST_DOMAIN=shopos.cluster.local' >> infra.env"
    sh "echo 'SPIRE_OIDC_URL=http://spiffe-oidc-discovery-provider.spire-system.svc.cluster.local' >> infra.env"
    echo 'SPIRE installed — HA server (3 replicas), SPIFFE CSI driver, OIDC discovery provider'
}
return this
