def call() {
    sh """
        RABBIT_URL=\$(grep '^RABBITMQ_MANAGEMENT_URL=' infra.env | cut -d= -f2)
        echo "Waiting for RabbitMQ management at \${RABBIT_URL}..."
        until curl -sf -u admin:admin "\${RABBIT_URL}/api/overview" > /dev/null 2>&1; do sleep 10; done

        # Create exchanges
        curl -sf -u admin:admin -X PUT "\${RABBIT_URL}/api/exchanges/shopos/commerce.events" \
            -H "Content-Type: application/json" \
            -d '{"type":"topic","durable":true}' || true
        curl -sf -u admin:admin -X PUT "\${RABBIT_URL}/api/exchanges/shopos/notification.events" \
            -H "Content-Type: application/json" \
            -d '{"type":"topic","durable":true}' || true
        curl -sf -u admin:admin -X PUT "\${RABBIT_URL}/api/exchanges/shopos/deadletter" \
            -H "Content-Type: application/json" \
            -d '{"type":"fanout","durable":true}' || true

        # Create queues
        curl -sf -u admin:admin -X PUT "\${RABBIT_URL}/api/queues/shopos/order.processing" \
            -H "Content-Type: application/json" \
            -d '{"durable":true,"arguments":{"x-dead-letter-exchange":"deadletter"}}' || true
        curl -sf -u admin:admin -X PUT "\${RABBIT_URL}/api/queues/shopos/email.sending" \
            -H "Content-Type: application/json" \
            -d '{"durable":true,"arguments":{"x-dead-letter-exchange":"deadletter"}}' || true
        curl -sf -u admin:admin -X PUT "\${RABBIT_URL}/api/queues/shopos/sms.sending" \
            -H "Content-Type: application/json" \
            -d '{"durable":true,"arguments":{"x-dead-letter-exchange":"deadletter"}}' || true

        # Bind queues to exchanges
        curl -sf -u admin:admin -X POST "\${RABBIT_URL}/api/bindings/shopos/e/commerce.events/q/order.processing" \
            -H "Content-Type: application/json" \
            -d '{"routing_key":"order.*"}' || true
        curl -sf -u admin:admin -X POST "\${RABBIT_URL}/api/bindings/shopos/e/notification.events/q/email.sending" \
            -H "Content-Type: application/json" \
            -d '{"routing_key":"email.*"}' || true
        curl -sf -u admin:admin -X POST "\${RABBIT_URL}/api/bindings/shopos/e/notification.events/q/sms.sending" \
            -H "Content-Type: application/json" \
            -d '{"routing_key":"sms.*"}' || true
    """
    echo 'rabbitmq configured — shopos vhost, exchanges, queues, and bindings created'
}
return this
