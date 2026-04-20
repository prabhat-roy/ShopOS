def call() {
    sh """
        helm upgrade --install pyrra observability/pyrra/charts             --namespace pyrra             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^PYRRA_/d' infra.env || true"
    sh "echo 'PYRRA_URL=http://pyrra-pyrra.pyrra.svc.cluster.local:9099' >> infra.env"
    echo 'pyrra installed'
}
return this
