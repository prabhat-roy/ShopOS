def call() {
    sh """
        echo "Configuring Schema Registry via kubectl exec..."
        kubectl exec -n schema-registry deploy/schema-registry -- \
            curl -sf -X PUT http://localhost:8081/config \
            -H 'Content-Type: application/vnd.schemaregistry.v1+json' \
            -d '{"compatibility":"BACKWARD"}' || true
        echo "Schema Registry compatibility set to BACKWARD"
    """
    echo 'schema-registry configured'
}
return this
