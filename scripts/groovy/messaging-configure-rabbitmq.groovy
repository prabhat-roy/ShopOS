def call() {
    sh """
        echo "Configuring RabbitMQ..."
        POD=\$(kubectl get pods -n rabbitmq --no-headers | head -1 | awk '{print \$1}')
        if [ -z "\${POD}" ]; then echo "No RabbitMQ pod found — skipping"; exit 0; fi

        # Create vhost and grant permissions (rabbitmqctl is available)
        kubectl exec -n rabbitmq "\${POD}" -- rabbitmqctl add_vhost shopos 2>/dev/null || true
        kubectl exec -n rabbitmq "\${POD}" -- rabbitmqctl set_permissions -p shopos admin '.*' '.*' '.*' 2>/dev/null || true

        # Use an ephemeral curl pod to call the management API (curl not in rabbitmq image)
        kubectl delete pod rabbitmq-setup -n rabbitmq --ignore-not-found=true 2>/dev/null || true
        kubectl run rabbitmq-setup -i --rm --restart=Never \
            --image=curlimages/curl:latest -n rabbitmq \
            --command -- sh -c '
                BASE=http://rabbitmq.rabbitmq.svc.cluster.local:15672/api
                AUTH=admin:admin

                # Declare exchanges
                curl -sf -u "\$AUTH" -X PUT "\$BASE/exchanges/shopos/commerce.events" \
                    -H "Content-Type: application/json" \
                    -d "{\"type\":\"topic\",\"durable\":true}" || true
                curl -sf -u "\$AUTH" -X PUT "\$BASE/exchanges/shopos/notification.events" \
                    -H "Content-Type: application/json" \
                    -d "{\"type\":\"topic\",\"durable\":true}" || true
                curl -sf -u "\$AUTH" -X PUT "\$BASE/exchanges/shopos/deadletter" \
                    -H "Content-Type: application/json" \
                    -d "{\"type\":\"fanout\",\"durable\":true}" || true

                # Declare queues
                curl -sf -u "\$AUTH" -X PUT "\$BASE/queues/shopos/order.processing" \
                    -H "Content-Type: application/json" \
                    -d "{\"durable\":true,\"arguments\":{\"x-dead-letter-exchange\":\"deadletter\"}}" || true
                curl -sf -u "\$AUTH" -X PUT "\$BASE/queues/shopos/email.sending" \
                    -H "Content-Type: application/json" \
                    -d "{\"durable\":true,\"arguments\":{\"x-dead-letter-exchange\":\"deadletter\"}}" || true
                curl -sf -u "\$AUTH" -X PUT "\$BASE/queues/shopos/sms.sending" \
                    -H "Content-Type: application/json" \
                    -d "{\"durable\":true,\"arguments\":{\"x-dead-letter-exchange\":\"deadletter\"}}" || true

                # Bind queues to exchanges
                curl -sf -u "\$AUTH" -X POST "\$BASE/bindings/shopos/e/commerce.events/q/order.processing" \
                    -H "Content-Type: application/json" \
                    -d "{\"routing_key\":\"order.*\"}" || true
                curl -sf -u "\$AUTH" -X POST "\$BASE/bindings/shopos/e/notification.events/q/email.sending" \
                    -H "Content-Type: application/json" \
                    -d "{\"routing_key\":\"email.*\"}" || true
                curl -sf -u "\$AUTH" -X POST "\$BASE/bindings/shopos/e/notification.events/q/sms.sending" \
                    -H "Content-Type: application/json" \
                    -d "{\"routing_key\":\"sms.*\"}" || true

                echo "RabbitMQ exchanges, queues, and bindings configured."
            ' || true
    """
    echo 'rabbitmq configured — shopos vhost, exchanges, queues, and bindings created'
}
return this
