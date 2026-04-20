def call() {
    sh '''
        echo "=== Configure Logstash ==="

        kubectl rollout status deploy/logstash -n logstash --timeout=120s || true

        ELASTICSEARCH_URL=$(grep '^ELASTICSEARCH_URL=' infra.env 2>/dev/null | cut -d= -f2 \
            || echo "http://elasticsearch.elasticsearch.svc.cluster.local:9200")
        ES_HOST=$(echo "$ELASTICSEARCH_URL" | sed 's|http://||;s|:.*||')
        ES_PORT=$(echo "$ELASTICSEARCH_URL" | sed 's|.*:||')

        # Patch the Logstash pipeline ConfigMap
        kubectl create configmap logstash-pipeline \
            -n logstash \
            --from-literal=logstash.conf="
input {
  beats {
    port => 5044
  }
  kafka {
    bootstrap_servers => \"kafka.kafka.svc.cluster.local:9092\"
    topics => [\"commerce.order.placed\", \"commerce.payment.processed\", \"security.fraud.detected\"]
    group_id => \"logstash-consumer\"
    codec => json
  }
}
filter {
  if [kubernetes] {
    mutate {
      add_field => { \"[@metadata][index]\" => \"shopos-k8s-logs\" }
    }
  } else {
    mutate {
      add_field => { \"[@metadata][index]\" => \"shopos-events\" }
    }
  }
}
output {
  elasticsearch {
    hosts => [\"${ES_HOST}:${ES_PORT}\"]
    index => \"%{[@metadata][index]}-%{+YYYY.MM.dd}\"
    action => \"index\"
  }
}" \
            --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null || true

        kubectl rollout restart deploy/logstash -n logstash 2>/dev/null || true

        echo "Logstash pipeline configured (Beats + Kafka → Elasticsearch)."
    '''
}
return this
