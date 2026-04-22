def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install rabbitmq messaging/rabbitmq/charts \
            --namespace rabbitmq \
            --create-namespace \
            --set fullnameOverride=rabbitmq \
            --set env.RABBITMQ_DEFAULT_USER=admin \
            --set env.RABBITMQ_DEFAULT_PASS=admin \
            --set env.RABBITMQ_DEFAULT_VHOST=shopos \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^RABBITMQ_/d' infra.env || true"
    sh "echo 'RABBITMQ_URL=amqp://rabbitmq.rabbitmq.svc.cluster.local:5672' >> infra.env"
    sh "echo 'RABBITMQ_MANAGEMENT_URL=http://rabbitmq.rabbitmq.svc.cluster.local:15672' >> infra.env"
    sh "echo 'RABBITMQ_USER=admin' >> infra.env"
    sh "echo 'RABBITMQ_PASSWORD=admin' >> infra.env"
    echo 'rabbitmq installed'
}
return this
