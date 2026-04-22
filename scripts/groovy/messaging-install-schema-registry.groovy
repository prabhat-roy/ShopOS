def call() {
    def sc = load('scripts/groovy/cloud-storage-class.groovy').call()
    sh """
        helm upgrade --install schema-registry messaging/schema-registry/charts \
            --namespace schema-registry \
            --create-namespace \
            --set fullnameOverride=schema-registry \
            --set env.SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS=kafka.kafka.svc.cluster.local:9092 \
            --set env.SCHEMA_REGISTRY_HOST_NAME=schema-registry \
            --set env.SCHEMA_REGISTRY_LISTENERS=http://0.0.0.0:8081 \
            --set persistence.storageClass=${sc} \
            --wait --timeout 5m
    """
    sh "sed -i '/^SCHEMA_REGISTRY_/d' infra.env || true"
    sh "echo 'SCHEMA_REGISTRY_URL=http://schema-registry.schema-registry.svc.cluster.local:8081' >> infra.env"
    echo 'schema-registry installed'
}
return this
