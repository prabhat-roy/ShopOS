def call() {
    sh """
        helm upgrade --install kafka-ui messaging/kafka-ui/charts \
            --namespace kafka-ui \
            --create-namespace \
            --set fullnameOverride=kafka-ui \
            --set env.KAFKA_CLUSTERS_0_NAME=shopos \
            --set env.KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS=kafka.kafka.svc.cluster.local:9092 \
            --set env.KAFKA_CLUSTERS_0_SCHEMAREGISTRY=http://schema-registry.schema-registry.svc.cluster.local:8081 \
            --set env.KAFKA_CLUSTERS_0_KAFKACONNECT_0_NAME=connect \
            --set env.KAFKA_CLUSTERS_0_KAFKACONNECT_0_ADDRESS=http://kafka-connect.kafka-connect.svc.cluster.local:8083 \
            --wait --timeout 5m
    """
    sh "sed -i '/^KAFKA_UI_/d' infra.env || true"
    sh "echo 'KAFKA_UI_URL=http://kafka-ui.kafka-ui.svc.cluster.local:8080' >> infra.env"
    echo 'kafka-ui installed'
}
return this
