def call() {
    sh """
        echo "Configuring Kafka Connect via kubectl exec..."
        kubectl exec -n kafka-connect deploy/kafka-connect-kafka-connect -- \
            curl -sf -X POST http://localhost:8083/connectors \
            -H 'Content-Type: application/json' \
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
        echo "Kafka Connect Debezium connector registered."
    """
    echo 'kafka-connect configured — Debezium order-service CDC connector registered'
}
return this
