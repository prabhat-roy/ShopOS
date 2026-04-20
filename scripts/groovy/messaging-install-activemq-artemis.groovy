def call() {
    sh """
        helm upgrade --install activemq-artemis messaging/activemq-artemis/charts \
            --namespace activemq-artemis \
            --create-namespace \
            --set env.AMQ_USER=admin \
            --set env.AMQ_PASSWORD=admin \
            --wait --timeout 5m
    """
    sh "sed -i '/^ACTIVEMQ_/d' infra.env || true"
    sh "sed -i '/^ACTIVEMQ_URL=/d' infra.env 2>/dev/null || true; echo 'ACTIVEMQ_URL=tcp://activemq-artemis-activemq-artemis.activemq-artemis.svc.cluster.local:61616' >> infra.env" 
    sh "sed -i '/^ACTIVEMQ_CONSOLE_URL=/d' infra.env 2>/dev/null || true; echo 'ACTIVEMQ_CONSOLE_URL=http://activemq-artemis-activemq-artemis.activemq-artemis.svc.cluster.local:8161' >> infra.env" 
    sh "sed -i '/^ACTIVEMQ_USER=/d' infra.env 2>/dev/null || true; echo 'ACTIVEMQ_USER=admin' >> infra.env" 
    sh "sed -i '/^ACTIVEMQ_PASSWORD=/d' infra.env 2>/dev/null || true; echo 'ACTIVEMQ_PASSWORD=admin' >> infra.env" 
    echo 'activemq-artemis installed'
}
return this
