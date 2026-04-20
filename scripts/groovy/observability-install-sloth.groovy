def call() {
    sh """
        helm upgrade --install sloth observability/sloth/charts             --namespace sloth             --create-namespace             --wait --timeout 5m
    """
    sh "sed -i '/^SLOTH_/d' infra.env || true"
    sh "echo 'SLOTH_URL=http://sloth-sloth.sloth.svc.cluster.local:8080' >> infra.env"
    echo 'sloth installed'
}
return this
