def call() {
    sh """
        KAFKA=kafka-kafka.kafka.svc.cluster.local:9092
        echo "Waiting for Kafka at \${KAFKA}..."
        until kubectl exec -n kafka deploy/kafka-kafka -- \
            kafka-topics --bootstrap-server localhost:9092 --list > /dev/null 2>&1; do sleep 10; done

        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic identity.user.registered --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic identity.user.deleted --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic commerce.order.placed --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic commerce.order.cancelled --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic commerce.order.fulfilled --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic commerce.payment.processed --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic commerce.payment.failed --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic commerce.cart.abandoned --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic supplychain.shipment.created --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic supplychain.shipment.updated --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic supplychain.inventory.low --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic supplychain.inventory.restocked --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic notification.email.requested --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic notification.sms.requested --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic notification.push.requested --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic analytics.page.viewed --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic analytics.product.clicked --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic analytics.search.performed --partitions 6 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic security.fraud.detected --partitions 3 --replication-factor 1
        kubectl exec -n kafka deploy/kafka-kafka -- kafka-topics \
            --bootstrap-server localhost:9092 --create --if-not-exists \
            --topic security.login.failed --partitions 3 --replication-factor 1

        echo "All ShopOS Kafka topics created."
    """
    echo 'kafka configured — all ShopOS topics created'
}
return this
