def call() {
    sh """
        helm upgrade --install rabbitmq messaging/rabbitmq/charts \
            --namespace rabbitmq \
            --create-namespace \
            --set env.RABBITMQ_DEFAULT_USER=admin \
            --set env.RABBITMQ_DEFAULT_PASS=admin \
            --set env.RABBITMQ_DEFAULT_VHOST=shopos \
            --wait --timeout 5m
    """
    sh "sed -i '/^RABBITMQ_/d' infra.env || true"
    sh "sed -i '/^RABBITMQ_URL=/d' infra.env 2>/dev/null || true; echo 'RABBITMQ_URL=amqp://rabbitmq-rabbitmq.rabbitmq.svc.cluster.local:5672' >> infra.env" 
    sh "sed -i '/^RABBITMQ_MANAGEMENT_URL=/d' infra.env 2>/dev/null || true; echo 'RABBITMQ_MANAGEMENT_URL=http://rabbitmq-rabbitmq.rabbitmq.svc.cluster.local:15672' >> infra.env" 
    sh "sed -i '/^RABBITMQ_USER=/d' infra.env 2>/dev/null || true; echo 'RABBITMQ_USER=admin' >> infra.env" 
    sh "sed -i '/^RABBITMQ_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'RABBITMQ_PASSWORD=admin' >> infra.env" 
    echo 'rabbitmq installed'
}
return this
