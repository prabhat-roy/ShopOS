def call() {
    sh '''
        helm upgrade --install pomerium security/pomerium/charts \
            --namespace pomerium \
            --create-namespace \
            --set config.rootDomain=shopos.local \
            --set config.generateTLS=true \
            --set config.insecure=true \
            --set authenticate.idp.provider=oidc \
            --set authenticate.idp.clientID=pomerium-client \
            --set authenticate.idp.clientSecret=pomerium-secret \
            --set "authenticate.idp.url=http://keycloak.keycloak.svc.cluster.local:8080/realms/shopos" \
            --set authenticate.idp.scopes=openid,profile,email,groups \
            --set authenticate.redirectUrl=https://authenticate.shopos.local/oauth2/callback \
            --set proxy.signingKey="" \
            --set databroker.storage.type=memory \
            --set operator.enabled=false \
            --set redis.enabled=false \
            --set replicaCount=2 \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=128Mi \
            --set metrics.enabled=true \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/pomerium -n pomerium --timeout=5m"
    sh "sed -i '/^POMERIUM_/d' infra.env || true"
    sh "sed -i '/^POMERIUM_URL=/d' infra.env 2>/dev/null || true; echo 'POMERIUM_URL=http://pomerium.pomerium.svc.cluster.local:80' >> infra.env" 
    sh "sed -i '/^POMERIUM_AUTHENTICATE_URL=/d' infra.env 2>/dev/null || true; echo 'POMERIUM_AUTHENTICATE_URL=https://authenticate.shopos.local' >> infra.env" 
    echo 'Pomerium installed — identity-aware proxy with Keycloak OIDC, zero-trust access'
}
return this
