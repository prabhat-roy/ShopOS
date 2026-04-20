def call() {
    sh """
        helm upgrade --install gimlet gitops/charts/gimlet \
            --namespace gimlet \
            --create-namespace \
            --wait --timeout 5m
    """
    sh "sed -i '/^GIMLET_/d' infra.env || true"
    sh "sed -i '/^GIMLET_URL=/d' infra.env 2>/dev/null || true; echo 'GIMLET_URL=http://gimlet-gimlet.gimlet.svc.cluster.local:9000' >> infra.env" 
    sh "sed -i '/^GIMLET_USER=/d' infra.env 2>/dev/null || true; echo 'GIMLET_USER=admin' >> infra.env" 
    sh "sed -i '/^GIMLET_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'GIMLET_PASSWORD=gimlet' >> infra.env" 
    echo 'gimlet installed'
}
return this
