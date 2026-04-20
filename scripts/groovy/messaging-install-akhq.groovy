def call() {
    sh """
        helm upgrade --install akhq messaging/akhq/charts \
            --namespace akhq \
            --create-namespace \
            --set env.MICRONAUT_CONFIG_FILES=/app/application.yml \
            --wait --timeout 5m
    """
    sh "sed -i '/^AKHQ_/d' infra.env || true"
    sh "sed -i '/^AKHQ_URL=/d' infra.env 2>/dev/null || true; echo 'AKHQ_URL=http://akhq-akhq.akhq.svc.cluster.local:8080' >> infra.env" 
    echo 'akhq installed'
}
return this
