def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install kafka messaging/kafka/charts \
            --namespace kafka \
            --create-namespace \
            --set fullnameOverride=kafka \
            --set env.KAFKA_BROKER_ID=1 \
            --set env.KAFKA_ZOOKEEPER_CONNECT=zookeeper.zookeeper.svc.cluster.local:2181 \
            --set env.KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
            --set env.KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka.kafka.svc.cluster.local:9092 \
            --set env.KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT \
            --set env.KAFKA_INTER_BROKER_LISTENER_NAME=PLAINTEXT \
            --set env.KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
            --set env.KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR=1 \
            --set env.KAFKA_TRANSACTION_STATE_LOG_MIN_ISR=1 \
            --set env.KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
            --set env.KAFKA_LOG_DIRS=/var/lib/data \
            --set persistence.storageClass=${sc} \
            --wait --timeout 15m
    """
    sh "sed -i '/^KAFKA_/d' infra.env || true"
    sh "echo 'KAFKA_URL=kafka.kafka.svc.cluster.local:9092' >> infra.env"
    echo 'kafka installed'
}
return this
