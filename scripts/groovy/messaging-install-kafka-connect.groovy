def call() {
    sh """
        helm upgrade --install kafka-connect messaging/kafka-connect/charts \
            --namespace kafka-connect \
            --create-namespace \
            --set env.CONNECT_BOOTSTRAP_SERVERS=kafka-kafka.kafka.svc.cluster.local:9092 \
            --set env.CONNECT_REST_PORT=8083 \
            --set env.CONNECT_GROUP_ID=kafka-connect \
            --set env.CONNECT_CONFIG_STORAGE_TOPIC=connect-configs \
            --set env.CONNECT_OFFSET_STORAGE_TOPIC=connect-offsets \
            --set env.CONNECT_STATUS_STORAGE_TOPIC=connect-status \
            --set env.CONNECT_KEY_CONVERTER=org.apache.kafka.connect.json.JsonConverter \
            --set env.CONNECT_VALUE_CONVERTER=org.apache.kafka.connect.json.JsonConverter \
            --wait --timeout 5m
    """
    sh "sed -i '/^KAFKA_CONNECT_/d' infra.env || true"
    sh "sed -i '/^KAFKA_CONNECT_URL=/d' infra.env 2>/dev/null || true; echo 'KAFKA_CONNECT_URL=http://kafka-connect-kafka-connect.kafka-connect.svc.cluster.local:8083' >> infra.env" 
    echo 'kafka-connect installed'
}
return this
