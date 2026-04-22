def call() {
    sh """
        helm upgrade --install kafka-connect messaging/kafka-connect/charts \
            --namespace kafka-connect \
            --create-namespace \
            --set fullnameOverride=kafka-connect \
            --set env.CONNECT_BOOTSTRAP_SERVERS=kafka.kafka.svc.cluster.local:9092 \
            --set env.CONNECT_REST_PORT=8083 \
            --set env.CONNECT_GROUP_ID=kafka-connect \
            --set env.CONNECT_CONFIG_STORAGE_TOPIC=connect-configs \
            --set env.CONNECT_OFFSET_STORAGE_TOPIC=connect-offsets \
            --set env.CONNECT_STATUS_STORAGE_TOPIC=connect-status \
            --set env.CONNECT_KEY_CONVERTER=org.apache.kafka.connect.json.JsonConverter \
            --set env.CONNECT_VALUE_CONVERTER=org.apache.kafka.connect.json.JsonConverter \
            --set env.CONNECT_REST_ADVERTISED_HOST_NAME=kafka-connect.kafka-connect.svc.cluster.local \
            --set env.CONNECT_CONFIG_STORAGE_REPLICATION_FACTOR=1 \
            --set env.CONNECT_OFFSET_STORAGE_REPLICATION_FACTOR=1 \
            --set env.CONNECT_STATUS_STORAGE_REPLICATION_FACTOR=1 \
            --set env.CONNECT_OFFSET_STORAGE_PARTITIONS=1 \
            --set env.CONNECT_STATUS_STORAGE_PARTITIONS=1 \
            --wait --timeout 15m
    """
    sh "sed -i '/^KAFKA_CONNECT_/d' infra.env || true"
    sh "echo 'KAFKA_CONNECT_URL=http://kafka-connect.kafka-connect.svc.cluster.local:8083' >> infra.env"
    echo 'kafka-connect installed'
}
return this
