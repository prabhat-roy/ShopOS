def call() {
    sh '''
        helm upgrade --install openfga security/openfga/charts \
            --namespace openfga \
            --create-namespace \
            --set image.tag=v1.5.7 \
            --set replicaCount=2 \
            --set resources.requests.cpu=100m \
            --set resources.requests.memory=256Mi \
            --set datastore.engine=postgres \
            --set datastore.uri=postgres://openfga:openfga@openfga-postgresql.openfga.svc.cluster.local:5432/openfga \
            --set postgresql.enabled=true \
            --set postgresql.auth.username=openfga \
            --set postgresql.auth.password=openfga \
            --set postgresql.auth.database=openfga \
            --set grpc.enabled=true \
            --set grpc.port=8081 \
            --set http.enabled=true \
            --set http.port=8080 \
            --set playground.enabled=true \
            --set playground.port=3000 \
            --set profiler.enabled=false \
            --set metrics.enabled=true \
            --set metrics.port=2112 \
            --set authn.method=none \
            --set maxTuplesPerWrite=100 \
            --set maxTypesPerAuthorizationModel=100 \
            --set changelogHorizonOffset=0 \
            --wait --timeout 5m
    '''
    sh "kubectl rollout status deployment/openfga -n openfga --timeout=5m"
    sh "sed -i '/^OPENFGA_/d' infra.env || true"
    sh "sed -i '/^OPENFGA_URL=/d' infra.env 2>/dev/null || true; echo 'OPENFGA_URL=http://openfga.openfga.svc.cluster.local:8080' >> infra.env" 
    sh "sed -i '/^OPENFGA_GRPC_URL=/d' infra.env 2>/dev/null || true; echo 'OPENFGA_GRPC_URL=openfga.openfga.svc.cluster.local:8081' >> infra.env" 
    sh "sed -i '/^OPENFGA_PLAYGROUND_URL=/d' infra.env 2>/dev/null || true; echo 'OPENFGA_PLAYGROUND_URL=http://openfga.openfga.svc.cluster.local:3000' >> infra.env" 
    echo 'OpenFGA installed — Zanzibar-based authorization with PostgreSQL, gRPC+HTTP, playground UI'
}
return this
