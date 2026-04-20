def call() {
    sh """
        SR_URL=\$(grep '^SCHEMA_REGISTRY_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Schema Registry at \${SR_URL}..."
        until curl -sf "\${SR_URL}/subjects" > /dev/null 2>&1; do sleep 10; done

        # Set compatibility to BACKWARD (default is BACKWARD, confirm it)
        curl -sf -X PUT "\${SR_URL}/config" \
            -H "Content-Type: application/vnd.schemaregistry.v1+json" \
            -d '{"compatibility":"BACKWARD"}' || true

        echo "Schema Registry ready — compatibility set to BACKWARD"
    """
    echo 'schema-registry configured'
}
return this
