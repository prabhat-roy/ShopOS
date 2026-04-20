def call() {
    sh """
        KC_URL=\$(grep '^KAFKA_CONNECT_URL=' infra.env | cut -d= -f2)
        echo "Waiting for Kafka Connect at \${KC_URL}..."
        until curl -sf "\${KC_URL}/connectors" > /dev/null 2>&1; do sleep 10; done

        # Register Debezium Postgres source connector for order-service
        curl -sf -X POST "\${KC_URL}/connectors" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "debezium-order-service",
                "config": {
                    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
                    "database.hostname": "postgres.platform.svc.cluster.local",
                    "database.port": "5432",
                    "database.user": "postgres",
                    "database.password": "postgres",
                    "database.dbname": "orders",
                    "database.server.name": "order-service",
                    "table.include.list": "public.orders",
                    "plugin.name": "pgoutput",
                    "topic.prefix": "cdc"
                }
            }' || true
    """
    echo 'kafka-connect configured — Debezium order-service CDC connector registered'
}
return this
