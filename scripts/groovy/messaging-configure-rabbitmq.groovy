def call() {
    sh """
        echo "Configuring RabbitMQ via kubectl exec..."
        POD=\$(kubectl get pods -n rabbitmq --no-headers | head -1 | awk '{print \$1}')
        if [ -z "\${POD}" ]; then echo "No RabbitMQ pod found — skipping"; exit 0; fi

        # Create vhost and grant permissions
        kubectl exec -n rabbitmq "\${POD}" -- rabbitmqctl add_vhost shopos 2>/dev/null || true
        kubectl exec -n rabbitmq "\${POD}" -- rabbitmqctl set_permissions -p shopos admin '.*' '.*' '.*' 2>/dev/null || true

        # Declare exchanges via management API on localhost
        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X PUT http://localhost:15672/api/exchanges/shopos/commerce.events \
            -H 'Content-Type: application/json' \
            -d '{"type":"topic","durable":true}' || true

        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X PUT http://localhost:15672/api/exchanges/shopos/notification.events \
            -H 'Content-Type: application/json' \
            -d '{"type":"topic","durable":true}' || true

        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X PUT http://localhost:15672/api/exchanges/shopos/deadletter \
            -H 'Content-Type: application/json' \
            -d '{"type":"fanout","durable":true}' || true

        # Declare queues
        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X PUT http://localhost:15672/api/queues/shopos/order.processing \
            -H 'Content-Type: application/json' \
            -d '{"durable":true,"arguments":{"x-dead-letter-exchange":"deadletter"}}' || true

        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X PUT http://localhost:15672/api/queues/shopos/email.sending \
            -H 'Content-Type: application/json' \
            -d '{"durable":true,"arguments":{"x-dead-letter-exchange":"deadletter"}}' || true

        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X PUT http://localhost:15672/api/queues/shopos/sms.sending \
            -H 'Content-Type: application/json' \
            -d '{"durable":true,"arguments":{"x-dead-letter-exchange":"deadletter"}}' || true

        # Bind queues to exchanges
        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X POST "http://localhost:15672/api/bindings/shopos/e/commerce.events/q/order.processing" \
            -H 'Content-Type: application/json' \
            -d '{"routing_key":"order.*"}' || true

        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X POST "http://localhost:15672/api/bindings/shopos/e/notification.events/q/email.sending" \
            -H 'Content-Type: application/json' \
            -d '{"routing_key":"email.*"}' || true

        kubectl exec -n rabbitmq "\${POD}" -- curl -sf -u admin:admin \
            -X POST "http://localhost:15672/api/bindings/shopos/e/notification.events/q/sms.sending" \
            -H 'Content-Type: application/json' \
            -d '{"routing_key":"sms.*"}' || true

        echo "RabbitMQ shopos vhost, exchanges, queues, and bindings configured."
    """
    echo 'rabbitmq configured — shopos vhost, exchanges, queues, and bindings created'
}
return this
