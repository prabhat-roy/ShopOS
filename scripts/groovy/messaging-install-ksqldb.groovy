def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install ksqldb messaging/ksqldb/charts \
            --namespace ksqldb \
            --create-namespace \
            --set fullnameOverride=ksqldb \
            --set env.KSQL_BOOTSTRAP_SERVERS=kafka.kafka.svc.cluster.local:9092 \
            --set env.KSQL_LISTENERS=http://0.0.0.0:8088 \
            --set env.KSQL_HOST_NAME=ksqldb.ksqldb.svc.cluster.local \
            --set env.KSQL_KSQL_SCHEMA_REGISTRY_URL=http://schema-registry.schema-registry.svc.cluster.local:8081 \
            --set persistence.storageClass=${sc} \
            --wait --timeout 10m
    """
    sh "sed -i '/^KSQLDB_/d' infra.env || true"
    sh "echo 'KSQLDB_URL=http://ksqldb.ksqldb.svc.cluster.local:8088' >> infra.env"
    echo 'ksqldb installed'
}
return this
