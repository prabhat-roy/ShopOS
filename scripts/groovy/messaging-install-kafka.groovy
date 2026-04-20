def call() {
    sh """
        helm upgrade --install kafka messaging/kafka/charts \
            --namespace kafka \
            --create-namespace \
            --set env.KAFKA_BROKER_ID=1 \
            --set env.KAFKA_ZOOKEEPER_CONNECT=zookeeper-zookeeper.zookeeper.svc.cluster.local:2181 \
            --set env.KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka-kafka.kafka.svc.cluster.local:9092 \
            --set env.KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
            --set env.KAFKA_AUTO_CREATE_TOPICS_ENABLE=true \
            --wait --timeout 5m
    """
    sh "sed -i '/^KAFKA_/d' infra.env || true"
    sh "echo 'KAFKA_URL=kafka-kafka.kafka.svc.cluster.local:9092' >> infra.env"
    echo 'kafka installed'
}
return this
