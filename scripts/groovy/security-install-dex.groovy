def call() {
    sh '''
        helm upgrade --install dex security/dex/charts \
            --namespace dex \
            --create-namespace \
            --set replicaCount=2 \
            --set image.tag=v2.41.0 \
            --set config.issuer=http://dex.dex.svc.cluster.local:5556 \
            --set config.storage.type=kubernetes \
            --set config.storage.config.inCluster=true \
            --set config.web.http=0.0.0.0:5556 \
            --set config.grpc.addr=0.0.0.0:5557 \
            --set config.grpc.reflection=true \
            --set config.oauth2.responseTypes[0]=code \
            --set config.oauth2.skipApprovalScreen=true \
            --set config.oauth2.alwaysShowLoginScreen=false \
            --set config.connectors[0].type=oidc \
            --set config.connectors[0].id=keycloak \
            --set config.connectors[0].name=Keycloak \
            --set "config.connectors[0].config.issuer=http://keycloak.keycloak.svc.cluster.local:8080/realms/shopos" \
            --set config.connectors[0].config.clientID=dex-client \
            --set config.connectors[0].config.clientSecret=dex-secret \
            --set config.connectors[0].config.scopes[0]=openid \
            --set config.connectors[0].config.scopes[1]=profile \
            --set config.connectors[0].config.scopes[2]=email \
            --set config.connectors[0].config.scopes[3]=groups \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=128Mi \
            --set serviceMonitor.enabled=false \
            --set metrics.enabled=true \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/dex -n dex --timeout=5m"
    sh "sed -i '/^DEX_/d' infra.env || true"
    sh "echo 'DEX_URL=http://dex.dex.svc.cluster.local:5556' >> infra.env"
    sh "echo 'DEX_GRPC_URL=dex.dex.svc.cluster.local:5557' >> infra.env"
    echo 'Dex OIDC federation installed — Kubernetes storage, Keycloak connector, gRPC API'
}
return this
