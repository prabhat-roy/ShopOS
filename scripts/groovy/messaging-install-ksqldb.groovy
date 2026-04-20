def call() {
    sh """
        helm upgrade --install ksqldb messaging/ksqldb/charts \
            --namespace ksqldb \
            --create-namespace \
            --set env.KSQL_BOOTSTRAP_SERVERS=kafka-kafka.kafka.svc.cluster.local:9092 \
            --set env.KSQL_LISTENERS=http://0.0.0.0:8088 \
            --set env.KSQL_KSQL_SCHEMA_REGISTRY_URL=http://schema-registry-schema-registry.schema-registry.svc.cluster.local:8081 \
            --wait --timeout 5m
    """
    sh "sed -i '/^KSQLDB_/d' infra.env || true"
    sh "echo 'KSQLDB_URL=http://ksqldb-ksqldb.ksqldb.svc.cluster.local:8088' >> infra.env"
    echo 'ksqldb installed'
}
return this
