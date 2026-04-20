def call() {
    sh '''
        helm upgrade --install fulcio security/fulcio/charts \
            --namespace sigstore-system \
            --create-namespace \
            --set server.replicaCount=2 \
            --set server.image.tag=v1.6.4 \
            --set server.args.grpcPort=5554 \
            --set server.args.httpPort=5555 \
            --set server.args.metricsPort=2112 \
            --set config.contents.OIDCIssuers.keycloak.IssuerURL=http://keycloak.keycloak.svc.cluster.local:8080/realms/sigstore \
            --set config.contents.OIDCIssuers.keycloak.ClientID=sigstore \
            --set config.contents.OIDCIssuers.keycloak.Type=kubernetes \
            --set config.contents.MetaIssuers."*.shopos.local".Type=kubernetes \
            --set server.resources.requests.cpu=100m \
            --set server.resources.requests.memory=128Mi \
            --set createcerts.enabled=true \
            --set ctlog.enabled=true \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/fulcio-server -n sigstore-system --timeout=5m"
    sh "sed -i '/^FULCIO_/d' infra.env || true"
    sh "echo 'FULCIO_URL=http://fulcio-server.sigstore-system.svc.cluster.local:5555' >> infra.env"
    sh "echo 'FULCIO_GRPC_URL=fulcio-server.sigstore-system.svc.cluster.local:5554' >> infra.env"
    echo 'Fulcio CA installed — Sigstore certificate authority with Keycloak OIDC and CT log'
}
return this
