def call() {
    sh """
        helm upgrade --install akhq messaging/akhq/charts \
            --namespace akhq \
            --create-namespace \
            --set fullnameOverride=akhq \
            --wait --timeout 5m
    """
    sh "sed -i '/^AKHQ_/d' infra.env || true"
    sh "echo 'AKHQ_URL=http://akhq.akhq.svc.cluster.local:8080' >> infra.env"
    echo 'akhq installed'
}
return this
