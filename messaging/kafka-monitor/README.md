# Kafka Monitor

## What is Kafka Monitor?

LinkedIn Kafka Monitor is an open-source tool that measures the end-to-end produce â†’ consume
latency and availability of a Kafka cluster by running a continuous synthetic workload.

It creates a dedicated monitoring topic (`kafka-monitor-topic`), produces messages to it at a
configurable rate, consumes those messages from the other end, and records the round-trip latency
along with availability metrics. This gives real signal about whether Kafka is healthy from the
perspective of a real producer-consumer pair â€” not just broker-level JMX stats.

---

## What it Measures

| Metric | Description |
|---|---|
| `produce-availability` | Fraction of produce attempts that succeeded |
| `consume-availability` | Fraction of messages consumed within the SLA window |
| `end-to-end latency (p50/p99/p999)` | Time from produce to consume, by percentile |
| `offset commit latency` | Time for consumer offset commits to succeed |
| `topic count` | Number of partitions and replicas on the monitor topic |

---

## Prometheus Integration (JMX Exporter)

Kafka Monitor exposes metrics via JMX. The JMX Prometheus Exporter sidecar scrapes these
and converts them to Prometheus format on port `5556`.

### Deployment snippet (abbreviated)

```yaml
containers:
  - name: kafka-monitor
    image: ghcr.io/linkedin/kafka-monitor:latest
    ports:
      - containerPort: 8778   # Jolokia / internal metrics
      - containerPort: 9999   # JMX

  - name: jmx-exporter
    image: bitnami/jmx-exporter:0.20.0
    args:
      - "5556"
      - /config/jmx-config.yaml
    ports:
      - containerPort: 5556   # Prometheus scrape port
    volumeMounts:
      - name: jmx-config
        mountPath: /config
```

### Prometheus scrape config

```yaml
- job_name: kafka-monitor
  static_configs:
    - targets: ['kafka-monitor.messaging.svc.cluster.local:5556']
  relabel_configs:
    - source_labels: [__address__]
      target_label: instance
```

### Grafana Dashboard

Import dashboard ID `14012` from Grafana.com for a pre-built Kafka Monitor overview panel
showing end-to-end latency trends, availability %, and alert thresholds.

---

## Alert Rules

```yaml
# PrometheusRule excerpt â€” add to observability/prometheus/rules/kafka-monitor.yaml
- alert: KafkaMonitorLowAvailability
  expr: kafka_monitor_consume_availability < 0.99
  for: 5m
  labels:
    severity: critical
    team: platform
  annotations:
    summary: "Kafka consumer availability below 99% for 5 minutes"

- alert: KafkaMonitorHighLatency
  expr: kafka_monitor_latency_ms{quantile="0.99"} > 1000
  for: 5m
  labels:
    severity: warning
    team: platform
  annotations:
    summary: "Kafka p99 end-to-end latency exceeds 1 second"
```

---

## Configuration Notes

- `topic.replication.factor: 3` â€” matches the ShopOS Kafka cluster's 3-broker setup.
- `produce.record.delay.ms: 100` â€” 10 messages/second; lightweight enough for production.
- `consume.latency.sla.ms: 1000` â€” messages not consumed within 1 second are counted as SLA
  violations in the `consume-availability` metric.
- The monitoring topic is auto-created if it doesn't exist (`topic.creation.enabled: true`).

---

## Running Kafka Monitor

```bash
# Via Docker (quick test)
docker run --rm \
  -e KAFKA_BOOTSTRAP_SERVERS=kafka:9092 \
  -e ZOOKEEPER_CONNECT=zookeeper:2181 \
  ghcr.io/linkedin/kafka-monitor:latest \
  /opt/kafka-monitor/bin/kafka-monitor-start.sh \
  /opt/kafka-monitor/config/kafka-monitor.yaml

# Via Kubernetes (deploy the Deployment)
kubectl apply -f messaging/kafka-monitor/kafka-monitor-deployment.yaml
```
