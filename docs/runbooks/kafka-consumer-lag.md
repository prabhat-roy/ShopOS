# Kafka Consumer Lag Runbook

> The `KafkaConsumerLagHigh` alert fires when any consumer group is more than 100k messages behind.

## Triage in 5 minutes
1. Identify the lagging consumer group:
   ```bash
   kubectl exec -n streaming kafka-0 -- bin/kafka-consumer-groups.sh \
     --bootstrap-server localhost:9092 --describe --group <consumergroup>
   ```
2. Check the consumer pod health: `kubectl get pods -n <ns> -l consumer-group=<consumergroup>`
3. Decide: is the consumer down (no progress) or slow (lag growing)?

## Consumer is down
- Roll the deployment: `kubectl rollout restart deploy/<svc> -n <ns>`
- Check pod logs for stack traces, OOM kills, or schema-registry validation errors
- If schema validation: the topic Avro schema in `events/<topic>.avsc` may be ahead of the consumer's compiled binding — pin the consumer or roll forward the consumer

## Consumer is slow
- Scale the consumer: `kubectl scale deploy/<svc> -n <ns> --replicas=<2x>`
- Check that partition count >= replica count (consumers idle if not)
   - To increase partitions: `kubectl exec kafka-0 -- bin/kafka-topics.sh --alter --topic <topic> --partitions <n>`
- Check downstream backpressure (e.g., Postgres slow on inserts) via Grafana Mimir

## Pause / resume a consumer
```bash
# Pause via Conduktor Gateway interceptor (preferred — does not lose offsets)
curl -X POST http://conduktor-gateway/admin/v1/intercept-config/pause-<group>
# Resume
curl -X DELETE http://conduktor-gateway/admin/v1/intercept-config/pause-<group>
```

## Drain a DLQ
```bash
kubectl create job dlq-drain-$(date +%s) \
  --from=cronjob/dlq-drain \
  --image-tag=latest \
  -n streaming \
  -- /app/drain --topic dlq.<group> --max 1000
```
