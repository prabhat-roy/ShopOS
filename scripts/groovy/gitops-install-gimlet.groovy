def call() {
    sh """
        helm upgrade --install gimlet gitops/charts/gimlet \
            --namespace gimlet \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^GIMLET_/d' infra.env || true"
    sh "echo 'GIMLET_URL=http://gimlet-gimlet.gimlet.svc.cluster.local:9000' >> infra.env"
    sh "echo 'GIMLET_USER=admin' >> infra.env"
    sh "echo 'GIMLET_PASSWORD=gimlet' >> infra.env"
    echo 'gimlet installed'
}
return this
